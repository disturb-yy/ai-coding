package golang

import (
	"fmt"
	"go/ast"

	"github.com/disturb-yy/codemap/internal/model"
)

// extractCallEdges 从 AST 文件提取函数调用边（跨模块 + 模块内）。
func extractCallEdges(file *ast.File, modulePath, receiverName string, moduleIndex map[string]*model.Module) []*model.CallEdge {
	var edges []*model.CallEdge

	localFuncs := make(map[string]bool)
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		name := fd.Name.Name
		if fd.Recv != nil && len(fd.Recv.List) > 0 {
			localFuncs[formatFuncName(receiverName, name)] = true
		} else {
			localFuncs[name] = true
		}
	}

	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Body == nil {
			continue
		}

		var caller string
		if fd.Recv != nil && len(fd.Recv.List) > 0 {
			caller = formatFuncName(receiverName, fd.Name.Name)
		} else {
			caller = fd.Name.Name
		}

		ast.Inspect(fd.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			edge := matchCallEdge(call, modulePath, caller, moduleIndex)
			if edge != nil {
				edges = append(edges, edge)
				return true
			}
			localEdge := matchLocalCallEdge(call, modulePath, caller, localFuncs)
			if localEdge != nil {
				edges = append(edges, localEdge)
			}
			return true
		})
	}

	return edges
}

func matchCallEdge(call *ast.CallExpr, modulePath, caller string, moduleIndex map[string]*model.Module) *model.CallEdge {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return nil
	}

	pkg := x.Name
	calleeModule := resolveModuleByPkg(pkg, moduleIndex)
	if calleeModule == "" {
		return nil
	}

	return &model.CallEdge{
		CallerModule: modulePath,
		CallerFunc:   caller,
		CalleeModule: calleeModule,
		CalleeFunc:   pkg + "." + sel.Sel.Name,
	}
}

func matchLocalCallEdge(call *ast.CallExpr, modulePath, caller string, localFuncs map[string]bool) *model.CallEdge {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return nil
	}
	callee := ident.Name
	if !localFuncs[callee] {
		return nil
	}
	return &model.CallEdge{
		CallerModule: modulePath,
		CallerFunc:   caller,
		CalleeModule: modulePath,
		CalleeFunc:   callee,
	}
}

func resolveModuleByPkg(pkg string, moduleIndex map[string]*model.Module) string {
	for _, mod := range moduleIndex {
		if mod.Name == pkg {
			return mod.Path
		}
	}
	return ""
}

func formatFuncName(receiver, funcName string) string {
	if receiver == "" {
		return funcName
	}
	return fmt.Sprintf("(*%s).%s", receiver, funcName)
}
