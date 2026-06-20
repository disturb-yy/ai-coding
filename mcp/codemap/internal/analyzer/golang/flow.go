package golang

import (
	"fmt"
	"go/ast"
	"sort"
	"strings"

	"github.com/disturb-yy/codemap/internal/model"
)

// extractFlows 从 AST 文件提取跨模块调用和数据流。
func extractFlows(file *ast.File, modulePath string, moduleIndex map[string]*model.Module) []*model.Flow {
	caller := &crossModuleVisitor{
		modulePath:  modulePath,
		moduleIndex: moduleIndex,
		internal:    make(map[string][]string),
		types:       make(map[string]bool),
	}
	ast.Walk(caller, file)

	var flows []*model.Flow

	for targetMod, fns := range caller.internal {
		sort.Strings(fns)
		flows = append(flows, &model.Flow{
			ID:      safeID(fmt.Sprintf("%s_to_%s", modulePath, targetMod)),
			Name:    fmt.Sprintf("%s calls %s", modulePath, targetMod),
			Trigger: modulePath,
			Steps:   buildSteps(modulePath, targetMod, fns),
		})
	}

	for t := range caller.types {
		flows = append(flows, &model.Flow{
			ID:      safeID(fmt.Sprintf("%s_to_type_%s", modulePath, t)),
			Name:    fmt.Sprintf("%s uses type %s", modulePath, t),
			Trigger: modulePath,
			Steps:   []string{fmt.Sprintf("%s references %s", modulePath, t)},
		})
	}

	return flows
}

func safeID(id string) string {
	id = strings.ReplaceAll(id, "/", "_")
	return id
}

type crossModuleVisitor struct {
	modulePath  string
	moduleIndex map[string]*model.Module
	internal    map[string][]string
	types       map[string]bool
}

func (v *crossModuleVisitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	if call, ok := n.(*ast.CallExpr); ok {
		v.visitCall(call)
	}
	if sel, ok := n.(*ast.SelectorExpr); ok {
		v.visitSelector(sel)
	}
	return v
}

func (v *crossModuleVisitor) visitCall(call *ast.CallExpr) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}
	pkg := x.Name
	targetMod := v.resolveModule(pkg)
	if targetMod == "" || targetMod == v.modulePath {
		return
	}
	ref := pkg + "." + sel.Sel.Name
	v.internal[targetMod] = appendUnique(v.internal[targetMod], ref)
}

func (v *crossModuleVisitor) visitSelector(sel *ast.SelectorExpr) {
	x, ok := sel.X.(*ast.Ident)
	if !ok {
		return
	}
	pkg := x.Name
	targetMod := v.resolveModule(pkg)
	if targetMod == "" || targetMod == v.modulePath {
		return
	}
	typeRef := pkg + "." + sel.Sel.Name
	v.types[typeRef] = true
}

func (v *crossModuleVisitor) resolveModule(pkg string) string {
	for _, mod := range v.moduleIndex {
		if mod.Name == pkg {
			return mod.Path
		}
	}
	return ""
}

func buildSteps(from, to string, fns []string) []string {
	steps := []string{fmt.Sprintf("%s -> %s", from, to)}
	for _, fn := range fns {
		steps = append(steps, fmt.Sprintf("  call: %s", fn))
	}
	return steps
}

func appendUnique(slice []string, s string) []string {
	for _, existing := range slice {
		if existing == s {
			return slice
		}
	}
	return append(slice, s)
}
