package golang

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/disturb-yy/codemap/internal/model"
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

func TestAnalyze_SkipsGeneratedAndDependencyDirs(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "go.mod", "module example.com/test\n\ngo 1.22\n")
	writeFile(t, root, "main.go", "package main\nfunc main() {}\n")

	vendorDir := filepath.Join(root, "vendor", "example.com", "dep")
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		t.Fatalf("mkdir vendor: %v", err)
	}
	writeFile(t, vendorDir, "dep.go", "package dep\nfunc Do() {}\n")

	codemapDir := filepath.Join(root, ".codemap", "generated")
	if err := os.MkdirAll(codemapDir, 0755); err != nil {
		t.Fatalf("mkdir .codemap: %v", err)
	}
	writeFile(t, codemapDir, "generated.go", "package generated\nfunc Do() {}\n")

	project, err := New().Analyze(context.Background(), root)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(project.Modules) != 1 {
		t.Fatalf("modules = %+v, want only root module", project.Modules)
	}
	if project.Modules[0].Path != "." {
		t.Fatalf("module path = %q, want .", project.Modules[0].Path)
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
	http.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {})
}
`)

	a := New()
	project, err := a.Analyze(context.Background(), root)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	if len(project.Routes) != 3 {
		t.Fatalf("expected 3 routes, got %d", len(project.Routes))
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
	if !hasGoRoute(project.Routes, "GET", "/ready") {
		t.Error("missing GET /ready route")
	}
}

func TestAnalyze_Routes_CommonGoRouters(t *testing.T) {
	root := t.TempDir()

	writeFile(t, root, "go.mod", "module example.com/api\n\ngo 1.22\n")
	writeFile(t, root, "main.go", `package main
import "net/http"

type Router struct{}
func (Router) Get(string, any) {}
func (Router) Post(string, any) {}
func (Router) MethodFunc(string, string, any) {}
func (Router) Register(string, string, any) {}
func (Router) Handle(string, string, any) {}
func (Router) Group(string) Router { return Router{} }
func (Router) PathPrefix(string) Router { return Router{} }
func (Router) Subrouter() Router { return Router{} }
func (Router) HandleFunc(string, any) Router { return Router{} }
func (Router) Methods(...string) Router { return Router{} }

func listUsers(w http.ResponseWriter, r *http.Request) {}
func createUser(w http.ResponseWriter, r *http.Request) {}
func deleteUser(w http.ResponseWriter, r *http.Request) {}
func getAdmin(w http.ResponseWriter, r *http.Request) {}
func updateUser(w http.ResponseWriter, r *http.Request) {}
func listTeams(w http.ResponseWriter, r *http.Request) {}
func patchTeam(w http.ResponseWriter, r *http.Request) {}

func main() {
	const usersPath = "/users"
	r := Router{}
	api := r.Group("/api")
	api.Get(usersPath, listUsers)
	api.Post(usersPath, createUser)
	r.MethodFunc(http.MethodDelete, "/users/{id}", deleteUser)
	r.Register("PUT", "/users/{id}", updateUser)
	r.HandleFunc("/teams", listTeams).Methods(http.MethodGet, "POST")
	r.Handle("PATCH", "/teams/{id}", patchTeam)

	admin := r.PathPrefix("/admin").Subrouter()
	admin.Get("/stats", getAdmin)
}
`)

	a := New()
	project, err := a.Analyze(context.Background(), root)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}

	tests := []struct {
		method string
		path   string
	}{
		{"GET", "/api/users"},
		{"POST", "/api/users"},
		{"DELETE", "/users/{id}"},
		{"PUT", "/users/{id}"},
		{"GET", "/teams"},
		{"POST", "/teams"},
		{"PATCH", "/teams/{id}"},
		{"GET", "/admin/stats"},
	}
	for _, tt := range tests {
		if !hasGoRoute(project.Routes, tt.method, tt.path) {
			t.Errorf("missing %s %s; routes=%v", tt.method, tt.path, goRouteStrings(project.Routes))
		}
	}
}

func hasGoRoute(routes []*model.Route, method, path string) bool {
	for _, r := range routes {
		if r.Method == method && r.Path == path {
			return true
		}
	}
	return false
}

func goRouteStrings(routes []*model.Route) []string {
	result := make([]string, 0, len(routes))
	for _, r := range routes {
		result = append(result, r.Method+" "+r.Path)
	}
	return result
}
