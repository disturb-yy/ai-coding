package golang

import (
	"bufio"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	baseanalyzer "github.com/disturb-yy/codemap/internal/analyzer"
	"github.com/disturb-yy/codemap/internal/model"
)

type Analyzer struct {
	modulePath string
}

func New() *Analyzer { return &Analyzer{} }

func (a *Analyzer) Analyze(ctx context.Context, root string) (*model.Project, error) {
	modPath, err := readModulePath(root)
	if err != nil {
		return nil, fmt.Errorf("read go.mod: %w", err)
	}
	a.modulePath = modPath

	project := &model.Project{
		Name: filepath.Base(root),
		Root: root,
	}

	moduleIndex := make(map[string]*model.Module)
	sqlUsers := make(map[string]bool)

	// Pass 1: 扫描所有文件，构建 module index + 模块依赖 + SQL 检测
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if baseanalyzer.ShouldSkipDir(info) {
			return filepath.SkipDir
		}
		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		return a.passOne(root, path, moduleIndex, sqlUsers)
	})
	if err != nil {
		return nil, err
	}

	for _, module := range moduleIndex {
		project.Modules = append(project.Modules, module)
	}

	// Pass 2: 用完整 moduleIndex 提取 routes/flows/call edges/exported symbols
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if baseanalyzer.ShouldSkipDir(info) {
			return filepath.SkipDir
		}
		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		return a.passTwo(root, path, moduleIndex, project)
	})
	if err != nil {
		return nil, err
	}

	// Pass 3: 检测内嵌持久化层，生成虚拟模块
	a.detectEmbeddedStorage(project, sqlUsers)

	return project, nil
}

func readModulePath(root string) (string, error) {
	modFile := filepath.Join(root, "go.mod")
	f, err := os.Open(modFile)
	if err != nil {
		return "", fmt.Errorf("open go.mod: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if after, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(after), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scan go.mod: %w", err)
	}
	return "", fmt.Errorf("no module declaration in go.mod")
}

func (a *Analyzer) passOne(projectRoot, filePath string, moduleIndex map[string]*model.Module, sqlUsers map[string]bool) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		return nil
	}

	modulePath := resolveModulePath(projectRoot, filePath)
	module, ok := moduleIndex[modulePath]
	if !ok {
		module = &model.Module{
			Name: file.Name.Name,
			Path: modulePath,
		}
		moduleIndex[modulePath] = module
	}

	imports := ExtractImports(file)
	for _, imp := range imports {
		if !IsInternalImport(imp, a.modulePath) {
			continue
		}
		AddDependency(module, NormalizeImport(imp, a.modulePath))
	}

	if HasSQLImport(imports) {
		sqlUsers[modulePath] = true
	}

	return nil
}

func (a *Analyzer) passTwo(projectRoot, filePath string, moduleIndex map[string]*model.Module, project *model.Project) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil
	}

	modulePath := resolveModulePath(projectRoot, filePath)
	receiverName := extractReceiverName(file)

	routes := extractRoutes(file, modulePath, receiverName)
	project.Routes = append(project.Routes, routes...)

	flows := extractFlows(file, modulePath, moduleIndex)
	project.Flows = append(project.Flows, flows...)

	edges := extractCallEdges(file, modulePath, receiverName, moduleIndex)
	project.CallEdges = append(project.CallEdges, edges...)

	if mod, ok := moduleIndex[modulePath]; ok && !strings.HasSuffix(filePath, "_test.go") {
		extractExports(file, mod)
	}

	return nil
}

func (a *Analyzer) detectEmbeddedStorage(project *model.Project, sqlUsers map[string]bool) {
	if len(sqlUsers) == 0 {
		return
	}

	for _, m := range project.Modules {
		if m.Name == "sqlite" || m.Name == "data" {
			return
		}
	}

	virtual := &model.Module{
		Name:         "data",
		Path:         "storage/sqlite",
		Dependencies: []string{},
	}
	project.Modules = append(project.Modules, virtual)

	for _, m := range project.Modules {
		if sqlUsers[m.Path] {
			AddDependency(m, "storage/sqlite")
		}
	}
}

// extractExports 从单个 Go 文件中提取导出的类型、函数、方法和接口。
func extractExports(file *ast.File, mod *model.Module) {
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				name := ts.Name.Name
				mod.ExportedTypes = appendStrIfNew(mod.ExportedTypes, name)
				if _, isIface := ts.Type.(*ast.InterfaceType); isIface {
					mod.KeyInterfaces = appendStrIfNew(mod.KeyInterfaces, name)
				}
			}
		case *ast.FuncDecl:
			if !d.Name.IsExported() {
				continue
			}
			if d.Recv == nil {
				// Standalone function
				mod.ExportedFunctions = appendStrIfNew(mod.ExportedFunctions, d.Name.Name)
			} else if len(d.Recv.List) > 0 {
				// Method: func (r *Receiver) Method()
				recvName := receiverNameFromField(d.Recv.List[0])
				if recvName != "" {
					method := recvName + "." + d.Name.Name
					mod.ExportedMethods = appendStrIfNew(mod.ExportedMethods, method)
				}
			}
		}
	}
}

// receiverNameFromField extracts the type name from a receiver field.
// Handles both *T and T forms.
func receiverNameFromField(field *ast.Field) string {
	if star, ok := field.Type.(*ast.StarExpr); ok {
		if ident, ok := star.X.(*ast.Ident); ok {
			return ident.Name
		}
	}
	if ident, ok := field.Type.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func appendStrIfNew(slice []string, s string) []string {
	for _, existing := range slice {
		if existing == s {
			return slice
		}
	}
	return append(slice, s)
}

func extractReceiverName(file *ast.File) string {
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv == nil || len(fd.Recv.List) == 0 {
			continue
		}
		if star, ok := fd.Recv.List[0].Type.(*ast.StarExpr); ok {
			if ident, ok := star.X.(*ast.Ident); ok {
				return ident.Name
			}
		}
		if ident, ok := fd.Recv.List[0].Type.(*ast.Ident); ok {
			return ident.Name
		}
	}
	return file.Name.Name
}

func resolveModulePath(projectRoot string, filePath string) string {
	relative, err := filepath.Rel(projectRoot, filepath.Dir(filePath))
	if err != nil {
		return filepath.Dir(filePath)
	}
	return filepath.ToSlash(relative)
}

func AddDependency(module *model.Module, dependency string) {
	for _, dep := range module.Dependencies {
		if dep == dependency {
			return
		}
	}
	module.Dependencies = append(module.Dependencies, dependency)
}

func HasSQLImport(imports []string) bool {
	for _, imp := range imports {
		if imp == "database/sql" || strings.HasPrefix(imp, "gorm.io/gorm") ||
			imp == "modernc.org/sqlite" {
			return true
		}
	}
	return false
}
