package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/disturb-yy/codemap/internal/model"
	"github.com/disturb-yy/codemap/internal/storage/sqlite"
	"github.com/disturb-yy/codemap/internal/workspace"
)

func TestWorkspaceListProjects(t *testing.T) {
	root := t.TempDir()
	createWorkspaceProject(t, root, "auth", nil, nil)
	createWorkspaceProject(t, root, "login", nil, nil)

	registry, err := workspace.New(root)
	if err != nil {
		t.Fatalf("workspace.New: %v", err)
	}
	defer registry.Close()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	registerWorkspaceTools(server, registry)

	result, err := callTool(server, "list_projects", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	var parsed struct {
		Projects []workspace.Project `json:"projects"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("parse list_projects: %v\n%s", err, text)
	}
	if len(parsed.Projects) != 2 {
		t.Fatalf("projects = %+v, want 2", parsed.Projects)
	}
	if parsed.Projects[0].Name != "auth" || parsed.Projects[1].Name != "login" {
		t.Fatalf("projects = %+v, want auth/login", parsed.Projects)
	}
}

func TestWorkspaceSearchRouteWithProject(t *testing.T) {
	root := t.TempDir()
	createWorkspaceProject(t, root, "auth", nil, []*model.Route{
		{Method: "POST", Path: "/login", Handler: "internal/auth.Login", Module: "internal/auth"},
	})
	createWorkspaceProject(t, root, "login", nil, []*model.Route{
		{Method: "GET", Path: "/sessions", Handler: "internal/session.List", Module: "internal/session"},
	})

	registry, err := workspace.New(root)
	if err != nil {
		t.Fatalf("workspace.New: %v", err)
	}
	defer registry.Close()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	registerWorkspaceTools(server, registry)

	result, err := callTool(server, "search_route", map[string]any{"project": "auth", "query": "/login"})
	if err != nil {
		t.Fatal(err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "[project: auth]") || !strings.Contains(text, "POST /login") {
		t.Fatalf("unexpected search_route result:\n%s", text)
	}
	if strings.Contains(text, "/sessions") {
		t.Fatalf("search_route leaked another project:\n%s", text)
	}
}

func TestWorkspaceSearchRouteInfersAndStripsProject(t *testing.T) {
	root := t.TempDir()
	createWorkspaceProject(t, root, "auth", nil, []*model.Route{
		{Method: "POST", Path: "/login", Handler: "internal/auth.Login", Module: "internal/auth"},
	})
	createWorkspaceProject(t, root, "login", nil, nil)

	registry, err := workspace.New(root)
	if err != nil {
		t.Fatalf("workspace.New: %v", err)
	}
	defer registry.Close()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	registerWorkspaceTools(server, registry)

	result, err := callTool(server, "search_route", map[string]any{"query": "auth /login"})
	if err != nil {
		t.Fatal(err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "[project: auth]") || !strings.Contains(text, "POST /login") {
		t.Fatalf("unexpected inferred search_route result:\n%s", text)
	}
}

func TestWorkspaceSearchRouteRequiresProjectWhenAmbiguous(t *testing.T) {
	root := t.TempDir()
	createWorkspaceProject(t, root, "auth", nil, nil)
	createWorkspaceProject(t, root, "login", nil, nil)

	registry, err := workspace.New(root)
	if err != nil {
		t.Fatalf("workspace.New: %v", err)
	}
	defer registry.Close()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	registerWorkspaceTools(server, registry)

	result, err := callTool(server, "search_route", map[string]any{"query": "/profile"})
	if err != nil {
		t.Fatal(err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if !strings.Contains(text, "project required") {
		t.Fatalf("expected project required error, got:\n%s", text)
	}
}

func createWorkspaceProject(t *testing.T, root, name string, modules []*model.Module, routes []*model.Route) {
	t.Helper()
	projectRoot := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Join(projectRoot, ".codemap"), 0755); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "go.mod"), []byte("module example.com/"+name+"\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	db, err := sqlite.Open(filepath.Join(projectRoot, ".codemap", "codemap.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := sqlite.NewRepository(db)
	for _, module := range modules {
		if err := repo.SaveModule(module); err != nil {
			t.Fatalf("save module: %v", err)
		}
	}
	for _, route := range routes {
		if err := repo.SaveRoute(route); err != nil {
			t.Fatalf("save route: %v", err)
		}
	}
}
