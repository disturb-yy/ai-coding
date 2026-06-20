package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/disturb-yy/codemap/internal/storage"
)

// Repository is the storage interface used by MCP tools.
type Repository = storage.Repository

// Serve starts the MCP server over stdio with graceful shutdown.
func Serve(repo Repository, projectName, projectRoot string) error {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "codemap",
		Version: "0.1.0",
	}, nil)

	registerTools(server, repo, projectName, projectRoot)
	registerResources(server, repo, projectRoot)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown on SIGTERM/SIGINT
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		cancel()
	}()

	// Wrap Run with panic recovery so a single tool panic doesn't
	// kill the entire server process, which causes "unsupported call"
	// on subsequent requests until Codex restarts the transport.
	return runWithRecovery(server, ctx)
}

// runWithRecovery wraps server.Run with a deferred panic handler that
// logs the stack trace and re-panics only on fatal transport errors.
func runWithRecovery(server *mcp.Server, ctx context.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("MCP server panic recovered: %v\n%s", r, debug.Stack())
			err = fmt.Errorf("server panic: %v", r)
		}
	}()
	return server.Run(ctx, &mcp.StdioTransport{})
}

// registerResources registers MCP resources for the project.
func registerResources(server *mcp.Server, repo Repository, projectRoot string) {
	server.AddResource(
		&mcp.Resource{
			URI:         "codemap://modules",
			Name:        "Project Modules",
			Description: "List of all modules in the project with their paths, dependencies, exported types, functions, and interfaces.",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			modules, err := repo.SearchModule("")
			if err != nil {
				return nil, fmt.Errorf("list modules: %w", err)
			}
			data, _ := json.MarshalIndent(modules, "", "  ")
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      "codemap://modules",
					MIMEType: "application/json",
					Text:     string(data),
				}},
			}, nil
		},
	)

	server.AddResourceTemplate(
		&mcp.ResourceTemplate{
			URITemplate: "codemap://module/{name}",
			Name:        "Module Detail",
			Description: "Details of a single module by name, including path, dependencies, exported types, functions, and interfaces.",
			MIMEType:    "application/json",
		},
		func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
			uri := req.Params.URI
			name := strings.TrimPrefix(uri, "codemap://module/")
			if name == "" || name == uri {
				return nil, fmt.Errorf("invalid resource URI %q", uri)
			}
			m, err := repo.FindModule(name)
			if err != nil {
				return nil, fmt.Errorf("find module: %w", err)
			}
			if m == nil {
				return nil, fmt.Errorf("module %q not found", name)
			}
			data, _ := json.MarshalIndent(m, "", "  ")
			return &mcp.ReadResourceResult{
				Contents: []*mcp.ResourceContents{{
					URI:      uri,
					MIMEType: "application/json",
					Text:     string(data),
				}},
			}, nil
		},
	)

	// Register .codemap/ Markdown documentation as MCP resources.
	codemapDir := filepath.Join(projectRoot, ".codemap")
	mdCategories := []struct {
		uriPrefix string
		dir       string
		desc      string
	}{
		{"codemap://architecture/", "architecture", "Architecture documentation"},
		{"codemap://callgraph/", "callgraph", "Call graph documentation"},
		{"codemap://flows/", "flows", "Data/call flow documentation"},
		{"codemap://modules-doc/", "modules", "Per-module Markdown documentation"},
		{"codemap://routes/", "routes", "HTTP route documentation"},
	}
	for _, cat := range mdCategories {
		cat := cat
		templateURI := cat.uriPrefix + "{name}"
		server.AddResourceTemplate(
			&mcp.ResourceTemplate{
				URITemplate: templateURI,
				Name:        cat.desc,
				Description: fmt.Sprintf("Markdown documentation from .codemap/%s/{name}.md", cat.dir),
				MIMEType:    "text/markdown",
			},
			func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				uri := req.Params.URI
				name := strings.TrimPrefix(uri, cat.uriPrefix)
				if name == "" || name == uri {
					return nil, fmt.Errorf("invalid resource URI %q", uri)
				}
				name = sanitizeResourceName(name)
				fpath := filepath.Join(codemapDir, cat.dir, name+".md")
				data, err := os.ReadFile(fpath)
				if err != nil {
					if os.IsNotExist(err) {
						return nil, fmt.Errorf("document %q not found in .codemap/%s/", name, cat.dir)
					}
					return nil, fmt.Errorf("read %s: %w", fpath, err)
				}
				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{{
						URI:      uri,
						MIMEType: "text/markdown",
						Text:     string(data),
					}},
				}, nil
			},
		)
	}
}

// sanitizeResourceName strips path traversal characters from a resource name.
func sanitizeResourceName(name string) string {
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	return name
}
