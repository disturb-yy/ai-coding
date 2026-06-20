package golang

import "testing"

func TestIsInternalImport(t *testing.T) {
	const modPath = "github.com/a/b"

	tests := []struct {
		name       string
		importPath string
		want       bool
	}{
		{"exact_match", "github.com/a/b", true},
		{"sub_package", "github.com/a/b/internal/order", true},
		{"deep_sub", "github.com/a/b/pkg/config", true},
		{"stdlib", "fmt", false},
		{"external", "github.com/gin-gonic/gin", false},
		{"same_prefix_diff_module", "github.com/a/ba", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInternalImport(tt.importPath, modPath)
			if got != tt.want {
				t.Errorf("IsInternalImport(%q, %q) = %v, want %v", tt.importPath, modPath, got, tt.want)
			}
		})
	}
}

func TestNormalizeImport(t *testing.T) {
	const modPath = "github.com/a/b"

	tests := []struct {
		name       string
		importPath string
		want       string
	}{
		{"exact_module", "github.com/a/b", "."},
		{"sub_package", "github.com/a/b/internal/order", "internal/order"},
		{"deep_sub", "github.com/a/b/pkg/config", "pkg/config"},
		{"stdlib", "fmt", "fmt"},
		{"external", "github.com/gin-gonic/gin", "github.com/gin-gonic/gin"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeImport(tt.importPath, modPath)
			if got != tt.want {
				t.Errorf("NormalizeImport(%q, %q) = %q, want %q", tt.importPath, modPath, got, tt.want)
			}
		})
	}
}
