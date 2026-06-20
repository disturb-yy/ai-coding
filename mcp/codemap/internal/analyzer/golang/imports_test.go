package golang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func parseImports(t *testing.T, src string) *ast.File {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return f
}

func TestExtractImports_Single(t *testing.T) {
	f := parseImports(t, `package p; import "fmt"`)
	got := ExtractImports(f)
	if len(got) != 1 || got[0] != "fmt" {
		t.Fatalf("expected [fmt], got %v", got)
	}
}

func TestExtractImports_Multiple(t *testing.T) {
	f := parseImports(t, `package p
import (
	"fmt"
	"os"
)`)
	got := ExtractImports(f)
	if len(got) != 2 {
		t.Fatalf("expected 2 imports, got %d: %v", len(got), got)
	}
}

func TestExtractImports_Empty(t *testing.T) {
	f := parseImports(t, `package p`)
	got := ExtractImports(f)
	if len(got) != 0 {
		t.Fatalf("expected 0 imports, got %v", got)
	}
}

func TestExtractImports_Aliased(t *testing.T) {
	f := parseImports(t, `package p; import x "pkg"`)
	got := ExtractImports(f)
	if len(got) != 1 || got[0] != "pkg" {
		t.Fatalf("expected [pkg], got %v", got)
	}
}
