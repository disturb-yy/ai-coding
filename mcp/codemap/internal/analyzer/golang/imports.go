package golang

import (
	"go/ast"
	"strings"
)

// ExtractImports 提取文件中的所有 import。
//
// 示例:
//
//	import (
//	    "fmt"
//	    "demo/internal/payment"
//	)
//
// 返回:
//
//	[]string{
//	    "fmt",
//	    "demo/internal/payment",
//	}
func ExtractImports(file *ast.File) []string {
	imports := make([]string, 0, len(file.Imports))

	for _, imp := range file.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		imports = append(imports, importPath)
	}

	return imports
}
