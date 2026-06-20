package java

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/disturb-yy/codemap/internal/model"
)

type Analyzer struct{}

func New() *Analyzer { return &Analyzer{} }

func (a *Analyzer) Analyze(ctx context.Context, root string) (*model.Project, error) {
	project := &model.Project{
		Name: filepath.Base(root),
		Root: root,
	}

	srcRoot := inferSourceRoot(root)
	if srcRoot == "" {
		return project, fmt.Errorf("no Java source directory found (expected src/main/java or src)")
	}

	moduleIndex := make(map[string]*model.Module)

	err := filepath.Walk(srcRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if info.IsDir() || filepath.Ext(path) != ".java" {
			return nil
		}
		return analyzeFile(path, srcRoot, moduleIndex)
	})
	if err != nil {
		return nil, err
	}

	for _, module := range moduleIndex {
		project.Modules = append(project.Modules, module)
	}

	return project, nil
}

func inferSourceRoot(root string) string {
	candidates := []string{
		filepath.Join(root, "src", "main", "java"),
		filepath.Join(root, "src"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			return c
		}
	}
	entries, _ := os.ReadDir(root)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".java") {
			return root
		}
	}
	return ""
}

func analyzeFile(filePath, srcRoot string, moduleIndex map[string]*model.Module) error {
	f, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var pkg, className string
	var classIsInterface bool
	var publicMethods []string
	var imports []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "/*") || strings.HasPrefix(line, "*") || strings.HasPrefix(line, "@") {
			continue
		}
		if after, ok := strings.CutPrefix(line, "package "); ok {
			pkg = strings.TrimSuffix(strings.TrimSpace(after), ";")
		}
		if after, ok := strings.CutPrefix(line, "import "); ok {
			imp := strings.TrimSuffix(strings.TrimSpace(after), ";")
			imports = append(imports, imp)
		}
		if name, isIface := extractPublicType(line); name != "" && className == "" {
			className = name
			classIsInterface = isIface
		}
		if method := extractPublicMethod(line); method != "" {
			publicMethods = append(publicMethods, method)
		}
	}

	// 从 .java 文件路径推导模块目录路径（如 com/example/service）
	modPath := modulePath(filePath, srcRoot)
	if modPath == "" {
		return nil
	}

	mod, ok := moduleIndex[modPath]
	if !ok {
		mod = &model.Module{Name: pkg, Path: modPath}
		if mod.Name == "" {
			mod.Name = filepath.Base(modPath)
		}
		moduleIndex[modPath] = mod
	}

	if className != "" {
		mod.ExportedTypes = appendStrIfNew(mod.ExportedTypes, className)
		if classIsInterface {
			mod.KeyInterfaces = appendStrIfNew(mod.KeyInterfaces, className)
		}
	}

	// Attach public methods to this module (ClassName.methodName format).
	if className != "" && len(publicMethods) > 0 {
		for _, method := range publicMethods {
			mod.ExportedMethods = appendStrIfNew(mod.ExportedMethods, className+"."+method)
		}
	}

	// 构建依赖：仅内部导入
	for _, imp := range imports {
		depModPath := importToModulePath(imp, pkg, srcRoot)
		if depModPath != "" && depModPath != modPath {
			addDepIfNew(mod, depModPath)
		}
	}

	return nil
}

func modulePath(filePath, srcRoot string) string {
	rel, err := filepath.Rel(srcRoot, filepath.Dir(filePath))
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

// importToModulePath 将 Java import 语句转换为项目内的模块路径。
// 仅处理内部导入（返回空字符串表示外部导入）。
func importToModulePath(imp, currentPkg, srcRoot string) string {
	// 跳过 JDK 和常见第三方库
	if isStdLib(imp) {
		return ""
	}

	// 移除末尾类名（大写开头）或通配符
	parts := strings.Split(imp, ".")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if last == "*" {
			parts = parts[:len(parts)-1]
		} else if len(last) > 0 && last[0] >= 'A' && last[0] <= 'Z' {
			parts = parts[:len(parts)-1]
		}
	}
	pkgPath := strings.Join(parts, ".")

	// 转换包路径为目录路径
	dirPath := strings.ReplaceAll(pkgPath, ".", "/")

	// 检查该目录是否存在于项目源码中
	fullPath := filepath.Join(srcRoot, dirPath)
	if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
		return dirPath
	}

	return ""
}

// isStdLib 判断是否为 JDK 标准库或常见第三方库导入。
func isStdLib(imp string) bool {
	prefixes := []string{
		"java.", "javax.", "jakarta.", "org.w3c.", "org.xml.",
		"org.springframework.", "org.hibernate.", "org.apache.",
		"org.junit.", "org.mockito.", "org.slf4j.", "org.testng.",
		"com.google.common.", "com.google.gson.",
		"com.fasterxml.", "lombok.", "io.swagger.",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(imp, p) {
			return true
		}
	}
	return false
}

// extractPublicType returns the type name and whether it's an interface.
func extractPublicType(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "@") {
		return "", false
	}
	for _, keyword := range []string{"class ", "interface ", "enum "} {
		if idx := strings.Index(line, "public "+keyword); idx >= 0 {
			rest := line[idx+len("public "+keyword):]
			name := takeIdent(strings.TrimSpace(rest))
			if name != "" {
				return name, keyword == "interface "
			}
		}
	}
	return "", false
}

// extractPublicMethod extracts a public method name from a single line.
// Matches: public ReturnType methodName(...), public static ReturnType methodName(...)
// Returns the bare method name (no class prefix).
func extractPublicMethod(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "@") || strings.HasPrefix(line, "//") {
		return ""
	}
	// Skip class/interface/enum declarations.
	for _, keyword := range []string{"class ", "interface ", "enum "} {
		if strings.Contains(line, "public "+keyword) {
			return ""
		}
	}
	// Match "public ... methodName("
	rest, ok := strings.CutPrefix(line, "public ")
	if !ok {
		return ""
	}
	// Strip optional modifiers.
	for _, mod := range []string{"static ", "abstract ", "default "} {
		rest = strings.TrimPrefix(rest, mod)
	}
	// Find the opening paren.
	parenIdx := strings.Index(rest, "(")
	if parenIdx < 0 {
		return ""
	}
	beforeParen := rest[:parenIdx]
	// The method name is the last identifier before the paren.
	parts := strings.Fields(beforeParen)
	if len(parts) == 0 {
		return ""
	}
	candidate := parts[len(parts)-1]
	// Filter out obvious non-method tokens.
	if candidate == "class" || candidate == "interface" || candidate == "enum" {
		return ""
	}
	// Must start with a letter or underscore.
	if len(candidate) > 0 && (isLetter(candidate[0]) || candidate[0] == '_') {
		return candidate
	}
	return ""
}

func takeIdent(s string) string {
	i := 0
	for i < len(s) && (isLetter(s[i]) || isDigit(s[i]) || s[i] == '_' || s[i] == '$') {
		i++
	}
	if i == 0 {
		return ""
	}
	return s[:i]
}

func isLetter(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }
func isDigit(c byte) bool  { return c >= '0' && c <= '9' }

func appendStrIfNew(slice []string, s string) []string {
	for _, existing := range slice {
		if existing == s {
			return slice
		}
	}
	return append(slice, s)
}

func addDepIfNew(mod *model.Module, dep string) {
	for _, existing := range mod.Dependencies {
		if existing == dep {
			return
		}
	}
	mod.Dependencies = append(mod.Dependencies, dep)
}
