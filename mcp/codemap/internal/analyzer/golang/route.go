package golang

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/disturb-yy/codemap/internal/model"
)

// httpMethods maps common router method names to HTTP methods.
var httpMethods = map[string]string{
	"GET": "GET", "POST": "POST", "PUT": "PUT",
	"DELETE": "DELETE", "PATCH": "PATCH", "HEAD": "HEAD",
	"OPTIONS": "OPTIONS", "CONNECT": "CONNECT", "TRACE": "TRACE",
	"Any": "ANY", "Handle": "ANY",
	"Get": "GET", "Post": "POST", "Put": "PUT",
	"Delete": "DELETE", "Patch": "PATCH", "Head": "HEAD",
	"Options": "OPTIONS", "Connect": "CONNECT", "Trace": "TRACE",
}

// extractRoutes 从 AST 文件提取 HTTP 路由。
func extractRoutes(file *ast.File, modulePath, receiverName string) []*model.Route {
	var routes []*model.Route
	ctx := collectRouteContext(file)

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		routes = append(routes, matchRouteCalls(call, modulePath, receiverName, ctx)...)
		return true
	})

	return dedupeRoutes(routes)
}

type routeContext struct {
	prefixes map[string]string
	strings  map[string]string
}

func matchRouteCalls(call *ast.CallExpr, modulePath, receiverName string, ctx routeContext) []*model.Route {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	method := sel.Sel.Name
	prefix := receiverPrefix(sel.X, ctx.prefixes)

	// gorilla/mux: r.HandleFunc(path, handler).Methods("GET", "POST")
	if method == "Methods" {
		return matchChainedMethods(call, modulePath, receiverName, ctx)
	}

	// go-restful/custom: route.GET("/path").To(handler)
	if method == "To" {
		return matchChainedTo(call, modulePath, receiverName, ctx)
	}

	// chi/gorilla/custom: Method("GET", "/path", handler), Handle(http.MethodPost, "/path", handler)
	if len(call.Args) >= 3 {
		if httpMethod := methodName(call.Args[0], ctx.strings); httpMethod != "" {
			path := stringValue(call.Args[1], ctx.strings)
			if path == "" {
				return nil
			}
			return []*model.Route{{
				Path:    joinRoutePath(prefix, path),
				Method:  httpMethod,
				Handler: handlerName(call.Args[len(call.Args)-1], modulePath, receiverName),
				Module:  modulePath,
			}}
		}
	}

	// net/http: HandleFunc / Handle
	if method == "HandleFunc" || method == "Handle" {
		if len(call.Args) < 2 {
			return nil
		}
		path := stringValue(call.Args[0], ctx.strings)
		if path == "" {
			return nil
		}
		httpMethod := method
		httpMethod, path = splitMethodPath(httpMethod, path)
		return []*model.Route{{
			Path:    joinRoutePath(prefix, path),
			Method:  httpMethod,
			Handler: handlerName(call.Args[1], modulePath, receiverName),
			Module:  modulePath,
		}}
	}

	// Gin/Echo/Fiber/Chi/HttpRouter/custom: GET/Get/Post("/path", handler)
	if httpMethod, ok := routeMethodFromSelector(method); ok {
		if len(call.Args) < 2 {
			return nil
		}
		path := stringValue(call.Args[0], ctx.strings)
		if path == "" {
			return nil
		}
		return []*model.Route{{
			Path:    joinRoutePath(prefix, path),
			Method:  httpMethod,
			Handler: handlerName(call.Args[len(call.Args)-1], modulePath, receiverName),
			Module:  modulePath,
		}}
	}

	return nil
}

func matchChainedTo(call *ast.CallExpr, modulePath, receiverName string, ctx routeContext) []*model.Route {
	if len(call.Args) == 0 {
		return nil
	}
	sel := call.Fun.(*ast.SelectorExpr)
	baseCall, ok := sel.X.(*ast.CallExpr)
	if !ok {
		return nil
	}
	baseSel, ok := baseCall.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	httpMethod, ok := routeMethodFromSelector(baseSel.Sel.Name)
	if !ok {
		return nil
	}
	if len(baseCall.Args) == 0 {
		return nil
	}
	path := stringValue(baseCall.Args[0], ctx.strings)
	if path == "" {
		return nil
	}
	return []*model.Route{{
		Path:    joinRoutePath(receiverPrefix(baseSel.X, ctx.prefixes), path),
		Method:  httpMethod,
		Handler: handlerName(call.Args[0], modulePath, receiverName),
		Module:  modulePath,
	}}
}

func matchChainedMethods(call *ast.CallExpr, modulePath, receiverName string, ctx routeContext) []*model.Route {
	sel := call.Fun.(*ast.SelectorExpr)
	baseCall, ok := sel.X.(*ast.CallExpr)
	if !ok {
		return nil
	}
	baseRoutes := matchRouteCalls(baseCall, modulePath, receiverName, ctx)
	if len(baseRoutes) == 0 {
		return nil
	}
	methods := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		if method := methodName(arg, ctx.strings); method != "" {
			methods = append(methods, method)
		}
	}
	if len(methods) == 0 {
		return nil
	}
	routes := make([]*model.Route, 0, len(baseRoutes)*len(methods))
	for _, base := range baseRoutes {
		for _, method := range methods {
			route := *base
			route.Method = method
			routes = append(routes, &route)
		}
	}
	return routes
}

func collectRouteContext(file *ast.File) routeContext {
	ctx := routeContext{
		prefixes: make(map[string]string),
		strings:  make(map[string]string),
	}
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			for i, lhs := range node.Lhs {
				if i >= len(node.Rhs) {
					continue
				}
				ident, ok := lhs.(*ast.Ident)
				if !ok {
					continue
				}
				if value := stringLiteral(node.Rhs[i]); value != "" {
					ctx.strings[ident.Name] = value
				}
				if prefix := routePrefixFromExpr(node.Rhs[i], ctx); prefix != "" {
					ctx.prefixes[ident.Name] = prefix
				}
			}
		case *ast.ValueSpec:
			for i, name := range node.Names {
				if i >= len(node.Values) {
					continue
				}
				if value := stringLiteral(node.Values[i]); value != "" {
					ctx.strings[name.Name] = value
				}
				if prefix := routePrefixFromExpr(node.Values[i], ctx); prefix != "" {
					ctx.prefixes[name.Name] = prefix
				}
			}
		}
		return true
	})
	return ctx
}

func routePrefixFromExpr(expr ast.Expr, ctx routeContext) string {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return ""
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	switch sel.Sel.Name {
	case "Group", "GroupPrefix", "Route", "Path", "PathPrefix", "Prefix":
		if len(call.Args) == 0 {
			return ""
		}
		path := stringValue(call.Args[0], ctx.strings)
		if path == "" {
			return ""
		}
		return joinRoutePath(receiverPrefix(sel.X, ctx.prefixes), path)
	case "Subrouter":
		return routePrefixFromExpr(sel.X, ctx)
	default:
		return ""
	}
}

func receiverPrefix(expr ast.Expr, prefixes map[string]string) string {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return prefixes[ident.Name]
}

func stringValue(expr ast.Expr, stringsByName map[string]string) string {
	if value := stringLiteral(expr); value != "" {
		return value
	}
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return stringsByName[ident.Name]
}

func stringLiteral(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return ""
	}
	return value
}

func methodName(expr ast.Expr, stringsByName map[string]string) string {
	if method := stringValue(expr, stringsByName); method != "" {
		return strings.ToUpper(method)
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	if method, ok := strings.CutPrefix(sel.Sel.Name, "Method"); ok {
		return strings.ToUpper(method)
	}
	return ""
}

func routeMethodFromSelector(method string) (string, bool) {
	if httpMethod, ok := httpMethods[method]; ok {
		return httpMethod, true
	}
	upper := strings.ToUpper(method)
	httpMethod, ok := httpMethods[upper]
	return httpMethod, ok
}

func splitMethodPath(fallback, pattern string) (string, string) {
	method, path, ok := strings.Cut(strings.TrimSpace(pattern), " ")
	if !ok || !strings.HasPrefix(path, "/") {
		return fallback, pattern
	}
	upper := strings.ToUpper(method)
	if _, ok := httpMethods[upper]; ok {
		return upper, path
	}
	return fallback, pattern
}

func joinRoutePath(base, child string) string {
	if base == "" {
		if strings.HasPrefix(child, "/") {
			return child
		}
		return "/" + child
	}
	return "/" + strings.Trim(strings.TrimSpace(base), "/") + "/" + strings.Trim(strings.TrimSpace(child), "/")
}

func dedupeRoutes(routes []*model.Route) []*model.Route {
	seen := make(map[string]*model.Route)
	hasExplicit := make(map[string]bool)
	for _, route := range routes {
		if route == nil {
			continue
		}
		key := route.Path + "\x00" + route.Handler + "\x00" + route.Module
		if route.Method != "Handle" && route.Method != "HandleFunc" {
			hasExplicit[key] = true
		}
	}

	result := make([]*model.Route, 0, len(routes))
	for _, route := range routes {
		if route == nil {
			continue
		}
		baseKey := route.Path + "\x00" + route.Handler + "\x00" + route.Module
		if hasExplicit[baseKey] && (route.Method == "Handle" || route.Method == "HandleFunc") {
			continue
		}
		key := route.Method + "\x00" + baseKey
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = route
		result = append(result, route)
	}
	return result
}

func handlerName(expr ast.Expr, modulePath, receiverName string) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return fmt.Sprintf("%s/%s.%s", modulePath, receiverName, e.Name)
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s/%s.%s", modulePath, receiverName, e.Sel.Name)
	case *ast.FuncLit:
		return fmt.Sprintf("%s/%s.anonymous", modulePath, receiverName)
	default:
		return fmt.Sprintf("%s/%s.unknown", modulePath, receiverName)
	}
}
