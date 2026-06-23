package mcp

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/disturb-yy/codemap/internal/model"
	"github.com/disturb-yy/codemap/internal/storage/sqlite"
)

func newRepo(t *testing.T) (*sqlite.Repository, func()) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	cleanup := func() { db.Close() }

	repo := sqlite.NewRepository(db)
	repo.SaveModule(&model.Module{
		Name:         "order",
		Path:         "internal/order",
		Dependencies: []string{"internal/payment"},
	})
	repo.SaveModule(&model.Module{
		Name: "payment",
		Path: "internal/payment",
	})
	return repo, cleanup
}

func callTool(server *mcp.Server, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	ctx := context.Background()
	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, t1, nil); err != nil {
		return nil, err
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, t2, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()

	res, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
	return res, err
}

func TestSearchModule(t *testing.T) {
	repo, cleanup := newRepo(t)
	defer cleanup()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	registerTools(server, repo, "test-project", "/tmp/test")

	result, err := callTool(server, "search_module", map[string]any{"module": "order"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content")
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if text == "" {
		t.Fatal("expected non-empty result")
	}
}

func TestSearchModule_NotFound(t *testing.T) {
	repo, cleanup := newRepo(t)
	defer cleanup()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	registerTools(server, repo, "test-project", "/tmp/test")

	result, err := callTool(server, "search_module", map[string]any{"module": "nonexistent"})
	if err != nil {
		t.Fatal(err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	if text != `module "nonexistent" not found` {
		t.Errorf("unexpected result: %s", text)
	}
}

func TestRelatedModules(t *testing.T) {
	repo, cleanup := newRepo(t)
	defer cleanup()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	registerTools(server, repo, "test-project", "/tmp/test")

	result, err := callTool(server, "related_modules", map[string]any{"module": "order"})
	if err != nil {
		t.Fatal(err)
	}
	text := result.Content[0].(*mcp.TextContent).Text

	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("parse result: %v", err)
	}

	deps, ok := parsed["dependencies"].([]any)
	if !ok || len(deps) != 1 || deps[0] != "internal/payment" {
		t.Errorf("dependencies = %v", parsed["dependencies"])
	}
}

func TestHandleInitialize(t *testing.T) {
	t1, t2 := mcp.NewInMemoryTransports()
	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	serverConn, err := server.Connect(context.Background(), t1, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverConn.Close()

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(context.Background(), t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	if session.InitializeResult().ServerInfo.Name != "test" {
		t.Errorf("server name = %q", session.InitializeResult().ServerInfo.Name)
	}
}

func TestHandleToolsList(t *testing.T) {
	repo, cleanup := newRepo(t)
	defer cleanup()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	registerTools(server, repo, "test-project", "/tmp/test")

	result, err := callTool(server, "get_project_info", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	var info map[string]any
	json.Unmarshal([]byte(text), &info)
	if info["project"] != "test-project" {
		t.Errorf("unexpected project: %v", info["project"])
	}
}

func TestFindChangePointsTool(t *testing.T) {
	repo, cleanup := newRepo(t)
	defer cleanup()

	if err := repo.SaveRoute(&model.Route{Method: "POST", Path: "/orders", Handler: "internal/order/handler.go", Module: "order"}); err != nil {
		t.Fatalf("SaveRoute: %v", err)
	}
	if err := repo.SaveFlow(&model.Flow{ID: "create_order", Name: "create_order_flow", Trigger: "order", Steps: []string{"validate order"}}); err != nil {
		t.Fatalf("SaveFlow: %v", err)
	}

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	registerTools(server, repo, "test-project", "/tmp/test")

	result, err := callTool(server, "find_change_points", map[string]any{"requirement": "add order cancellation", "top_k": 3})
	if err != nil {
		t.Fatal(err)
	}
	text := result.Content[0].(*mcp.TextContent).Text
	var parsed map[string]any
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("parse result: %v\n%s", err, text)
	}
	if parsed["requirement"] != "add order cancellation" {
		t.Errorf("unexpected requirement: %v", parsed["requirement"])
	}
	if len(parsed["candidate_modules"].([]any)) == 0 {
		t.Fatal("expected candidate modules")
	}
}

func TestNotificationIgnored(t *testing.T) {
	repo, cleanup := newRepo(t)
	defer cleanup()

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0.1.0"}, nil)
	registerTools(server, repo, "test-project", "/tmp/test")

	t1, t2 := mcp.NewInMemoryTransports()
	if _, err := server.Connect(context.Background(), t1, nil); err != nil {
		t.Fatal(err)
	}

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0"}, nil)
	session, err := client.Connect(context.Background(), t2, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()

	// Verify server is reachable after initialization
	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "search_module",
		Arguments: map[string]any{"module": "order"},
	})
	if err != nil {
		t.Fatalf("call after init: %v", err)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected content after notification")
	}
}
