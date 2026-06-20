package golang

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/disturb-yy/codemap/internal/model"
)

// httpMethods maps Gin-style method names to HTTP methods.
var httpMethods = map[string]string{
	"GET": "GET", "POST": "POST", "PUT": "PUT",
	"DELETE": "DELETE", "PATCH": "PATCH", "HEAD": "HEAD",
	"OPTIONS": "OPTIONS", "Any": "ANY", "Handle": "ANY",
}

// extractRoutes 从 AST 文件提取 HTTP 路由。
func extractRoutes(file *ast.File, modulePath, receiverName string) []*model.Route {
	var routes []*model.Route

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		r := matchRouteCall(call, modulePath, receiverName)
		if r != nil {
			routes = append(routes, r)
		}
		return true
	})

	return routes
}

func matchRouteCall(call *ast.CallExpr, modulePath, receiverName string) *model.Route {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	method := sel.Sel.Name

	// net/http: HandleFunc / Handle
	if method == "HandleFunc" || method == "Handle" {
		if len(call.Args) < 2 {
			return nil
		}
		path := stringLiteral(call.Args[0])
		if path == "" {
			return nil
		}
		return &model.Route{
			Path:    path,
			Method:  method,
			Handler: handlerName(call.Args[1], modulePath, receiverName),
			Module:  modulePath,
		}
	}

	// Gin: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, Any, Handle
	if httpMethod, ok := httpMethods[method]; ok {
		if len(call.Args) < 2 {
			return nil
		}
		path := stringLiteral(call.Args[0])
		if path == "" {
			return nil
		}
		return &model.Route{
			Path:    path,
			Method:  httpMethod,
			Handler: handlerName(call.Args[len(call.Args)-1], modulePath, receiverName),
			Module:  modulePath,
		}
	}

	return nil
}

func stringLiteral(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	return lit.Value[1 : len(lit.Value)-1] // strip quotes
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
