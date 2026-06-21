package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/disturb-yy/codemap/internal/model"
)

// registerTools registers all MCP tools on the server.
func registerTools(server *mcp.Server, repo Repository, projectName, projectRoot string) {
	registerSearchModule(server, repo)
	registerRelatedModules(server, repo)
	registerSearchRoute(server, repo)
	registerSearchFlow(server, repo)
	registerCallGraph(server, repo)
	registerImpactAnalysis(server, repo)
	registerGetProjectInfo(server, repo, projectName, projectRoot)
	registerListModules(server, repo)
	registerGetFeatureMap(server, repo)
	registerGetNavigationHints(server, repo)
}

func registerSearchModule(server *mcp.Server, repo Repository) {
	server.AddTool(
		&mcp.Tool{
			Name:        "search_module",
			Description: "Search for a module by name. Returns its path and dependencies.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"module":{"type":"string","description":"Module name to search for. Empty returns all modules."}},"required":["module"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct{ Module string }
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			results, err := repo.SearchModule(args.Module)
			if err != nil {
				return errorResult("search_failed", "search module: "+err.Error(), ""), nil
			}
			if len(results) == 0 {
				return textResult(fmt.Sprintf("module %q not found", args.Module)), nil
			}
			if len(results) == 1 {
				m := results[0]
				data, _ := json.MarshalIndent(map[string]any{
					"path": m.Path, "dependencies": m.Dependencies,
				}, "", "  ")
				return textResult(string(data)), nil
			}
			var list []map[string]any
			for _, m := range results {
				list = append(list, map[string]any{
					"name": m.Name, "path": m.Path, "dependencies": m.Dependencies,
				})
			}
			data, _ := json.MarshalIndent(map[string]any{
				"matched": len(results),
				"modules": list,
			}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerRelatedModules(server *mcp.Server, repo Repository) {
	server.AddTool(
		&mcp.Tool{
			Name:        "related_modules",
			Description: "List modules that depend on or are depended on by the given module.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"module":{"type":"string","description":"Module name to query."}},"required":["module"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct{ Module string }
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			m, err := repo.FindModule(args.Module)
			if err != nil {
				return errorResult("find_failed", "find module: "+err.Error(), ""), nil
			}
			if m == nil {
				return textResult(fmt.Sprintf("module %q not found", args.Module)), nil
			}
			all, err := repo.SearchModule("")
			if err != nil {
				return errorResult("list_failed", "list modules: "+err.Error(), ""), nil
			}
			var dependents []string
			for _, mod := range all {
				for _, dep := range mod.Dependencies {
					if dep == m.Path {
						dependents = append(dependents, mod.Name)
					}
				}
			}
			data, _ := json.MarshalIndent(map[string]any{
				"dependencies": m.Dependencies, "dependents": dependents,
			}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerSearchRoute(server *mcp.Server, repo Repository) {
	server.AddTool(
		&mcp.Tool{
			Name:        "search_route",
			Description: "Search for HTTP routes by path, method, or module.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","description":"Search query (path fragment, module name)."}},"required":["query"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct{ Query string }
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			if args.Query == "" {
				return textResult("query required — provide a path fragment or module name"), nil
			}
			routes, err := repo.FindRoutes(args.Query)
			if err != nil {
				return errorResult("find_routes_failed", "find routes: "+err.Error(), ""), nil
			}
			if len(routes) == 0 {
				return textResult(fmt.Sprintf("no routes matching %q", args.Query)), nil
			}
			var b strings.Builder
			const maxRoutes = 40
			totalRoutes := len(routes)
			r := routes
			if len(r) > maxRoutes {
				fmt.Fprintf(&b, "%d routes total (showing first %d):\n", totalRoutes, maxRoutes)
				r = r[:maxRoutes]
			}
			for _, rt := range r {
				fmt.Fprintf(&b, "%s %s → %s [%s]\n", rt.Method, rt.Path, rt.Handler, rt.Module)
			}
			return textResult(b.String()), nil
		}),
	)
}

func registerSearchFlow(server *mcp.Server, repo Repository) {
	server.AddTool(
		&mcp.Tool{
			Name:        "search_flow",
			Description: "Search for data/call flows by name or trigger.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","description":"Search query (flow name or trigger module)."}},"required":["query"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct{ Query string }
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			if args.Query == "" {
				return textResult("query required — provide a flow name fragment or trigger module"), nil
			}
			flows, err := repo.SearchFlow(args.Query)
			if err != nil {
				return errorResult("search_flow_failed", "search flow: "+err.Error(), ""), nil
			}
			if len(flows) == 0 {
				return textResult(fmt.Sprintf("no flows matching %q", args.Query)), nil
			}
			var b strings.Builder
			const maxFlows = 60
			total := len(flows)
			f := flows
			if len(f) > maxFlows {
				fmt.Fprintf(&b, "%d flows total (showing first %d):\n", total, maxFlows)
				f = f[:maxFlows]
			}
			for _, fl := range f {
				fmt.Fprintf(&b, "%s [%s]\n", fl.Name, fl.Trigger)
				for _, step := range fl.Steps {
					fmt.Fprintf(&b, "  └ %s\n", step)
				}
			}
			return textResult(b.String()), nil
		}),
	)
}

func registerCallGraph(server *mcp.Server, repo Repository) {
	server.AddTool(
		&mcp.Tool{
			Name:        "call_graph",
			Description: "Get the call graph for a module — which functions it calls.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"module":{"type":"string","description":"Module name to query."}},"required":["module"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct{ Module string }
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			lookupPath := resolveModulePath(repo, args.Module)
			edges, err := callWithRetry(func() ([]*model.CallEdge, error) {
				return repo.FindCallees(lookupPath)
			})
			if err != nil {
				return errorResult("find_callees_failed", "find callees: "+err.Error(), ""), nil
			}
			if len(edges) == 0 {
				return textResult(fmt.Sprintf("no call edges for %q", args.Module)), nil
			}
			var b strings.Builder
			for _, e := range edges {
				fmt.Fprintf(&b, "%s.%s → %s.%s\n", e.CallerModule, e.CallerFunc, e.CalleeModule, e.CalleeFunc)
			}
			return textResult(b.String()), nil
		}),
	)
}

func registerImpactAnalysis(server *mcp.Server, repo Repository) {
	server.AddTool(
		&mcp.Tool{
			Name:        "impact_analysis",
			Description: "Find all callers of a function — what would break if changed.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"function":{"type":"string","description":"Function name to analyze (partial match)."}},"required":["function"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct{ Function string }
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			edges, err := callWithRetry(func() ([]*model.CallEdge, error) {
				return repo.FindCallers(args.Function)
			})
			if err != nil {
				return errorResult("find_callers_failed", "find callers: "+err.Error(), ""), nil
			}
			if len(edges) == 0 {
				return textResult(fmt.Sprintf("no callers found for %q", args.Function)), nil
			}
			var b strings.Builder
			fmt.Fprintf(&b, "Impact of changing %q:\n", args.Function)
			for _, e := range edges {
				fmt.Fprintf(&b, "  %s.%s calls %s.%s\n", e.CallerModule, e.CallerFunc, e.CalleeModule, e.CalleeFunc)
			}
			return textResult(b.String()), nil
		}),
	)
}

func registerGetProjectInfo(server *mcp.Server, repo Repository, projectName, projectRoot string) {
	server.AddTool(
		&mcp.Tool{
			Name:        "get_project_info",
			Description: "Get project metadata: name, root path, and module count.",
			InputSchema: json.RawMessage(`{"type":"object"}`),
		},
		safeHandler(func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			modules, err := repo.SearchModule("")
			if err != nil {
				return errorResult("list_failed", "list modules: "+err.Error(), ""), nil
			}
			data, _ := json.MarshalIndent(map[string]any{
				"project":      projectName,
				"root":         projectRoot,
				"module_count": len(modules),
			}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerListModules(server *mcp.Server, repo Repository) {
	server.AddTool(
		&mcp.Tool{
			Name:        "list_modules",
			Description: "List all modules with their paths, dependencies, exported types, functions, methods, and key interfaces.",
			InputSchema: json.RawMessage(`{"type":"object"}`),
		},
		safeHandler(func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			modules, err := repo.SearchModule("")
			if err != nil {
				return errorResult("list_failed", "list modules: "+err.Error(), ""), nil
			}
			data, _ := json.MarshalIndent(modules, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerGetFeatureMap(server *mcp.Server, repo Repository) {
	server.AddTool(
		&mcp.Tool{
			Name:        "get_feature_map",
			Description: "Get the project's business feature map — features with their modules, routes, and flows.",
			InputSchema: json.RawMessage(`{"type":"object"}`),
		},
		safeHandler(func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			features, err := repo.GetFeatureMap()
			if err != nil {
				return errorResult("feature_map_failed", "get feature map: "+err.Error(), ""), nil
			}
			data, _ := json.MarshalIndent(map[string]any{"features": features}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerGetNavigationHints(server *mcp.Server, repo Repository) {
	server.AddTool(
		&mcp.Tool{
			Name:        "get_navigation_hints",
			Description: "Get project navigation guidance — entry files, related modules, flows, and risk areas per feature.",
			InputSchema: json.RawMessage(`{"type":"object"}`),
		},
		safeHandler(func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			hints, err := repo.GetNavigationHints()
			if err != nil {
				return errorResult("navigation_hints_failed", "get navigation hints: "+err.Error(), ""), nil
			}
			data, _ := json.MarshalIndent(map[string]any{"features": hints}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

// safeHandler wraps a tool handler with panic recovery.
func safeHandler(fn func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error)) func(context.Context, *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req *mcp.CallToolRequest) (result *mcp.CallToolResult, err error) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				err = fmt.Errorf("tool panic: %v\n%s", r, stack)
			}
		}()
		return fn(ctx, req)
	}
}

// callWithRetry executes a CallEdge-returning function with exponential backoff.
func callWithRetry(fn func() ([]*model.CallEdge, error)) ([]*model.CallEdge, error) {
	const maxRetries = 3
	var lastErr error
	for i := range maxRetries {
		edges, err := fn()
		if err == nil {
			return edges, nil
		}
		lastErr = err
		if i < maxRetries-1 {
			time.Sleep(time.Duration(1<<uint(i)) * 100 * time.Millisecond)
		}
	}
	return nil, fmt.Errorf("after %d retries: %w", maxRetries, lastErr)
}

// resolveModulePath converts a module name to its filesystem path.
func resolveModulePath(repo Repository, nameOrPath string) string {
	if strings.Contains(nameOrPath, "/") {
		return nameOrPath
	}
	m, err := repo.FindModule(nameOrPath)
	if err != nil || m == nil {
		return nameOrPath
	}
	return m.Path
}

// errorResult returns a structured JSON error as a text result.
func errorResult(code, reason, retryAfter string) *mcp.CallToolResult {
	data, _ := json.MarshalIndent(map[string]any{
		"error_code":  code,
		"reason":      reason,
		"retry_after": retryAfter,
	}, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}
}

// textResult wraps text in a CallToolResult.
func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
