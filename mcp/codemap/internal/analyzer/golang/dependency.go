package golang

import "strings"

// IsInternalImport 根据项目 module path 判断 import 是否属于项目内部依赖。
//
// modulePath 来自 go.mod 的 module 声明（如 github.com/disturb-yy/codemap）。
// 将 import 路径与 modulePath 匹配：完全匹配或以 modulePath/ 为前缀。
func IsInternalImport(importPath, modulePath string) bool {
	if importPath == modulePath {
		return true
	}
	if strings.HasPrefix(importPath, modulePath+"/") {
		return true
	}
	return false
}

// NormalizeImport 将 import 路径相对于 modulePath 归一化为项目内路径。
//
// 示例:
//
//	github.com/company/demo/internal/payment
//
// =>
//
//	internal/payment
func NormalizeImport(importPath, modulePath string) string {
	if importPath == modulePath {
		return "."
	}
	prefix := modulePath + "/"
	if strings.HasPrefix(importPath, prefix) {
		return importPath[len(prefix):]
	}
	return importPath
}
