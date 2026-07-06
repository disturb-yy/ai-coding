package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/disturb-yy/codemap/internal/cognitive"
	"github.com/disturb-yy/codemap/internal/model"
	"github.com/disturb-yy/codemap/internal/workspace"
)

const projectArgDescription = "Workspace child project/service name, e.g. auth or login. Use this when the user mentions a service or repository name."

// ServeWorkspace starts an MCP server that routes tool calls to child projects.
func ServeWorkspace(registry *workspace.Registry) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "codemap-workspace",
		Version: "0.1.0",
	}, nil)

	registerWorkspaceTools(server, registry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return runWithRecovery(server, ctx)
}

func registerWorkspaceTools(server *mcp.Server, registry *workspace.Registry) {
	registerListProjects(server, registry)
	registerWorkspaceGetProjectInfo(server, registry)
	registerWorkspaceListModules(server, registry)
	registerWorkspaceSearchModule(server, registry)
	registerWorkspaceRelatedModules(server, registry)
	registerWorkspaceSearchRoute(server, registry)
	registerWorkspaceSearchFlow(server, registry)
	registerWorkspaceCallGraph(server, registry)
	registerWorkspaceImpactAnalysis(server, registry)
	registerWorkspaceGetFeatureMap(server, registry)
	registerWorkspaceGetNavigationHints(server, registry)
	registerWorkspaceFindChangePoints(server, registry)
}

func registerListProjects(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "list_projects",
			Description: "List child projects available in the current workspace. Use this before querying when the target service/repository is unclear.",
			InputSchema: json.RawMessage(`{"type":"object"}`),
		},
		safeHandler(func(_ context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			data, _ := json.MarshalIndent(map[string]any{
				"workspace_root": registry.Root(),
				"projects":       registry.Projects(),
			}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerWorkspaceGetProjectInfo(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "get_project_info",
			Description: "Get workspace metadata, or child project metadata when project is provided.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"}}}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct{ Project string }
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			if strings.TrimSpace(args.Project) == "" {
				data, _ := json.MarshalIndent(map[string]any{
					"workspace_root": registry.Root(),
					"project_count":  len(registry.Projects()),
					"projects":       registry.Projects(),
				}, "", "  ")
				return textResult(string(data)), nil
			}
			project, repo, err := registry.Resolve(args.Project, "")
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			modules, err := repo.SearchModule("")
			if err != nil {
				return errorResult("list_failed", "list modules: "+err.Error(), ""), nil
			}
			data, _ := json.MarshalIndent(map[string]any{
				"project":      project.Name,
				"root":         project.Root,
				"module_count": len(modules),
			}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerWorkspaceListModules(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "list_modules",
			Description: "List modules for a workspace child project. If the user mentions a service/repository name such as auth or login, pass it as project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"}}}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct{ Project string }
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			project, repo, err := registry.Resolve(args.Project, "")
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			modules, err := repo.SearchModule("")
			if err != nil {
				return errorResult("list_failed", "list modules: "+err.Error(), ""), nil
			}
			data, _ := json.MarshalIndent(map[string]any{
				"project": project.Name,
				"root":    project.Root,
				"modules": modules,
			}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerWorkspaceSearchModule(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "search_module",
			Description: "Search for a module in a workspace child project. If the user mentions a service/repository name such as auth or login, pass it as project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"},"module":{"type":"string","description":"Module name to search for. Empty returns all modules."}},"required":["module"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Project string
				Module  string
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			project, repo, err := registry.Resolve(args.Project, args.Module)
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			moduleQuery := stripProjectName(args.Module, project.Name)
			results, err := repo.SearchModule(moduleQuery)
			if err != nil {
				return errorResult("search_failed", "search module: "+err.Error(), ""), nil
			}
			if len(results) == 0 {
				return textResult(fmt.Sprintf("module %q not found in project %q", moduleQuery, project.Name)), nil
			}
			data, _ := json.MarshalIndent(map[string]any{
				"project": project.Name,
				"root":    project.Root,
				"matched": len(results),
				"modules": results,
			}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerWorkspaceRelatedModules(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "related_modules",
			Description: "List modules that depend on or are depended on by the given module in a workspace child project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"},"module":{"type":"string","description":"Module name to query."}},"required":["module"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Project string
				Module  string
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			project, repo, err := registry.Resolve(args.Project, args.Module)
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			m, err := repo.FindModule(args.Module)
			if err != nil {
				return errorResult("find_failed", "find module: "+err.Error(), ""), nil
			}
			if m == nil {
				return textResult(fmt.Sprintf("module %q not found in project %q", args.Module, project.Name)), nil
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
				"project":      project.Name,
				"module":       m.Name,
				"dependencies": m.Dependencies,
				"dependents":   dependents,
			}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerWorkspaceSearchRoute(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "search_route",
			Description: "Search for HTTP routes in a workspace child project. If the user mentions a service/repository name such as auth or login, pass it as project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"},"query":{"type":"string","description":"Search query (path fragment, module name)."}},"required":["query"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Project string
				Query   string
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			if args.Query == "" {
				return textResult("query required; provide a path fragment or module name"), nil
			}
			project, repo, err := registry.Resolve(args.Project, args.Query)
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			query := stripProjectName(args.Query, project.Name)
			routes, err := repo.FindRoutes(query)
			if err != nil {
				return errorResult("find_routes_failed", "find routes: "+err.Error(), ""), nil
			}
			if len(routes) == 0 {
				return textResult(fmt.Sprintf("no routes matching %q in project %q", query, project.Name)), nil
			}
			return textResult(formatWorkspaceRoutes(project.Name, routes)), nil
		}),
	)
}

func registerWorkspaceSearchFlow(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "search_flow",
			Description: "Search for data/call flows in a workspace child project. If the user mentions a service/repository name such as auth or login, pass it as project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"},"query":{"type":"string","description":"Search query (flow name or trigger module)."}},"required":["query"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Project string
				Query   string
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			if args.Query == "" {
				return textResult("query required; provide a flow name fragment or trigger module"), nil
			}
			project, repo, err := registry.Resolve(args.Project, args.Query)
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			query := stripProjectName(args.Query, project.Name)
			flows, err := repo.SearchFlow(query)
			if err != nil {
				return errorResult("search_flow_failed", "search flow: "+err.Error(), ""), nil
			}
			if len(flows) == 0 {
				return textResult(fmt.Sprintf("no flows matching %q in project %q", query, project.Name)), nil
			}
			return textResult(formatWorkspaceFlows(project.Name, flows)), nil
		}),
	)
}

func registerWorkspaceCallGraph(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "call_graph",
			Description: "Get the call graph for a module in a workspace child project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"},"module":{"type":"string","description":"Module name to query."}},"required":["module"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Project string
				Module  string
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			project, repo, err := registry.Resolve(args.Project, args.Module)
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			lookupPath := resolveModulePath(repo, args.Module)
			edges, err := callWithRetry(func() ([]*model.CallEdge, error) {
				return repo.FindCallees(lookupPath)
			})
			if err != nil {
				return errorResult("find_callees_failed", "find callees: "+err.Error(), ""), nil
			}
			if len(edges) == 0 {
				return textResult(fmt.Sprintf("no call edges for %q in project %q", args.Module, project.Name)), nil
			}
			var b strings.Builder
			fmt.Fprintf(&b, "[project: %s]\n", project.Name)
			for _, e := range edges {
				fmt.Fprintf(&b, "%s.%s -> %s.%s\n", e.CallerModule, e.CallerFunc, e.CalleeModule, e.CalleeFunc)
			}
			return textResult(b.String()), nil
		}),
	)
}

func registerWorkspaceImpactAnalysis(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "impact_analysis",
			Description: "Find all callers of a function in a workspace child project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"},"function":{"type":"string","description":"Function name to analyze (partial match)."}},"required":["function"]}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Project  string
				Function string
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			project, repo, err := registry.Resolve(args.Project, args.Function)
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			edges, err := callWithRetry(func() ([]*model.CallEdge, error) {
				return repo.FindCallers(args.Function)
			})
			if err != nil {
				return errorResult("find_callers_failed", "find callers: "+err.Error(), ""), nil
			}
			if len(edges) == 0 {
				return textResult(fmt.Sprintf("no callers found for %q in project %q", args.Function, project.Name)), nil
			}
			var b strings.Builder
			fmt.Fprintf(&b, "[project: %s]\nImpact of changing %q:\n", project.Name, args.Function)
			for _, e := range edges {
				fmt.Fprintf(&b, "  %s.%s calls %s.%s\n", e.CallerModule, e.CallerFunc, e.CalleeModule, e.CalleeFunc)
			}
			return textResult(b.String()), nil
		}),
	)
}

func registerWorkspaceGetFeatureMap(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "get_feature_map",
			Description: "Get the business feature map for a workspace child project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"}}}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct{ Project string }
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			project, repo, err := registry.Resolve(args.Project, "")
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			features, err := repo.GetFeatureMap()
			if err != nil {
				return errorResult("feature_map_failed", "get feature map: "+err.Error(), ""), nil
			}
			data, _ := json.MarshalIndent(map[string]any{"project": project.Name, "features": features}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerWorkspaceGetNavigationHints(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "get_navigation_hints",
			Description: "Get navigation guidance for a workspace child project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"}}}`),
		},
		safeHandler(func(_ context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct{ Project string }
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			project, repo, err := registry.Resolve(args.Project, "")
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			hints, err := repo.GetNavigationHints()
			if err != nil {
				return errorResult("navigation_hints_failed", "get navigation hints: "+err.Error(), ""), nil
			}
			data, _ := json.MarshalIndent(map[string]any{"project": project.Name, "features": hints}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func registerWorkspaceFindChangePoints(server *mcp.Server, registry *workspace.Registry) {
	server.AddTool(
		&mcp.Tool{
			Name:        "find_change_points",
			Description: "Find likely modules, files, routes, flows and risks in a workspace child project. If the user mentions a service/repository name such as auth or login, pass it as project.",
			InputSchema: json.RawMessage(`{"type":"object","properties":{"project":{"type":"string","description":"` + projectArgDescription + `"},"requirement":{"type":"string","description":"User requirement text."},"top_k":{"type":"integer","description":"Maximum number of candidates to return."}},"required":["requirement"]}`),
		},
		safeHandler(func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			var args struct {
				Project     string `json:"project,omitempty"`
				Requirement string `json:"requirement"`
				TopK        int    `json:"top_k,omitempty"`
			}
			if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
				return nil, err
			}
			project, repo, err := registry.Resolve(args.Project, args.Requirement)
			if err != nil {
				return workspaceErrorResult(err), nil
			}
			result, err := cognitive.NewService(repo).FindChangePoints(ctx, cognitive.FindChangePointsRequest{
				Requirement: args.Requirement,
				TopK:        args.TopK,
			})
			if err != nil {
				return errorResult("find_change_points_failed", "find change points: "+err.Error(), ""), nil
			}
			data, _ := json.MarshalIndent(map[string]any{
				"project": project.Name,
				"root":    project.Root,
				"result":  result,
			}, "", "  ")
			return textResult(string(data)), nil
		}),
	)
}

func workspaceErrorResult(err error) *mcp.CallToolResult {
	return errorResult("workspace_project_required", err.Error(), "Call list_projects, then retry with the project argument.")
}

func stripProjectName(query, projectName string) string {
	query = strings.TrimSpace(strings.ReplaceAll(query, projectName, ""))
	if query == "" {
		return projectName
	}
	return strings.Join(strings.Fields(query), " ")
}

func formatWorkspaceRoutes(project string, routes []*model.Route) string {
	var b strings.Builder
	const maxRoutes = 40
	totalRoutes := len(routes)
	fmt.Fprintf(&b, "[project: %s]\n", project)
	r := routes
	if len(r) > maxRoutes {
		fmt.Fprintf(&b, "%d routes total (showing first %d):\n", totalRoutes, maxRoutes)
		r = r[:maxRoutes]
	}
	for _, rt := range r {
		fmt.Fprintf(&b, "%s %s -> %s [%s]\n", rt.Method, rt.Path, rt.Handler, rt.Module)
	}
	return b.String()
}

func formatWorkspaceFlows(project string, flows []*model.Flow) string {
	var b strings.Builder
	const maxFlows = 60
	total := len(flows)
	fmt.Fprintf(&b, "[project: %s]\n", project)
	f := flows
	if len(f) > maxFlows {
		fmt.Fprintf(&b, "%d flows total (showing first %d):\n", total, maxFlows)
		f = f[:maxFlows]
	}
	for _, fl := range f {
		fmt.Fprintf(&b, "%s [%s]\n", fl.Name, fl.Trigger)
		for _, step := range fl.Steps {
			fmt.Fprintf(&b, "  - %s\n", step)
		}
	}
	return b.String()
}
