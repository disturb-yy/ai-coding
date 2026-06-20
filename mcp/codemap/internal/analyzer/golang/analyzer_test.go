package golang

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestResolveModulePath(t *testing.T) {
	tests := []struct {
		name     string
		root     string
		filePath string
		want     string
	}{
		{"subdir", "/proj", "/proj/internal/order/svc.go", "internal/order"},
		{"root_go", "/proj", "/proj/main.go", "."},
		{"nested", "/proj", "/proj/cmd/server/main.go", "cmd/server"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveModulePath(tt.root, tt.filePath)
			if got != tt.want {
				t.Errorf("resolveModulePath(%q, %q) = %q, want %q",
					tt.root, tt.filePath, got, tt.want)
			}
		})
	}
}

func TestAnalyze(t *testing.T) {
	// 创建临时项目
	root := t.TempDir()

	writeFile(t, root, "go.mod", "module example.com/test\n\ngo 1.22\n")
	writeFile(t, root, "main.go", `package main
import "example.com/test/internal/order"
func main() { order.Do() }
`)

	orderDir := filepath.Join(root, "internal", "order")
	os.MkdirAll(orderDir, 0755)
	writeFile(t, orderDir, "order.go", "package order\nfunc Do() {}\n")

	payDir := filepath.Join(root, "internal", "payment")
	os.MkdirAll(payDir, 0755)
	writeFile(t, payDir, "pay.go", `package payment
import "example.com/test/internal/order"
func Pay() { order.Do() }
`)

	a := New()
	project, err := a.Analyze(context.Background(), root)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if len(project.Modules) != 3 {
		t.Fatalf("expected 3 modules, got %d", len(project.Modules))
	}

	// 按 path 排序以便确定性断言
	sort.Slice(project.Modules, func(i, j int) bool {
		return project.Modules[i].Path < project.Modules[j].Path
	})

	mainMod := project.Modules[0]
	if mainMod.Path != "." {
		t.Errorf("main path = %q, want .", mainMod.Path)
	}
	if len(mainMod.Dependencies) != 1 || mainMod.Dependencies[0] != "internal/order" {
		t.Errorf("main deps = %v, want [internal/order]", mainMod.Dependencies)
	}

	orderMod := project.Modules[1]
	if orderMod.Path != "internal/order" {
		t.Errorf("order path = %q, want internal/order", orderMod.Path)
	}
	if len(orderMod.Dependencies) != 0 {
		t.Errorf("order deps = %v, want []", orderMod.Dependencies)
	}

	payMod := project.Modules[2]
	if payMod.Path != "internal/payment" {
		t.Errorf("payment path = %q, want internal/payment", payMod.Path)
	}
	if len(payMod.Dependencies) != 1 || payMod.Dependencies[0] != "internal/order" {
		t.Errorf("payment deps = %v, want [internal/order]", payMod.Dependencies)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", name, err)
	}
}

func TestAnalyze_Routes(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "go.mod", "module example.com/api\n\ngo 1.22\n")
	writeFile(t, root, "main.go", `package main
import "net/http"
func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {})
	http.Handle("/api/data", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
}
`)

	a := New()
	project, err := a.Analyze(context.Background(), root)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if len(project.Routes) != 2 {
		t.Fatalf("expected 2 routes, got %d", len(project.Routes))
	}

	paths := make(map[string]bool)
	for _, r := range project.Routes {
		paths[r.Path] = true
	}
	if !paths["/health"] {
		t.Error("missing /health route")
	}
	if !paths["/api/data"] {
		t.Error("missing /api/data route")
	}
}
