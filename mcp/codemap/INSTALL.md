# CodeMap — Installation & Usage Guide

CodeMap is an agent-oriented project knowledge layer. It indexes Go codebases into a SQLite database
and exposes structured project knowledge (modules, dependencies, call graphs, flows, routes)
through the MCP protocol.

---

## 1. Build from Source

**Prerequisites**: Go 1.23+

```bash
git clone https://github.com/disturb-yy/codemap.git
cd codemap
go build -o codemap ./cmd/codemap/
```

Move the binary to your PATH:

```bash
sudo mv codemap /usr/local/bin/
# or, for personal use:
mkdir -p ~/.local/bin && mv codemap ~/.local/bin/
```

---

## 2. Index a Project

Run inside any Go project root:

```bash
cd /path/to/your-go-project
codemap
```

This produces:

```
your-project/
├── .codemap/
│   ├── codemap.db          # SQLite database (source of truth)
│   ├── INDEX.md            # Project entry point
│   ├── modules/            # Per-module Markdown docs
│   ├── architecture/       # Overview + dependency graph
│   ├── routes/             # HTTP route docs
│   ├── flows/              # Cross-module call flows
│   └── callgraph/          # Function-level call graph
```

A `.gitignore` entry for `.codemap/` is added automatically.

> Re-run `codemap` after significant code changes to keep the index up to date.

---

## 3. MCP Server Configuration

CodeMap exposes two kinds of MCP interfaces:

| Channel | Method | What It Provides |
|---------|--------|-----------------|
| **Tools** | `tools/call` | 7 query tools (search, impact analysis, call graph, etc.) |
| **Resources** | `resources/read` | 6 resource templates (Markdown docs, JSON module data) |

### 3.1 Codex / Codex++

Add to `~/.codex/config.toml` or Codex++ MCP settings:

```toml
[mcp_servers.codemap]
command = "/usr/local/bin/codemap"
args = ["-project", "/absolute/path/to/your-project", "--serve"]
```

### 3.2 Cursor

Add to `.cursor/mcp.json` in your project root:

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "${workspaceFolder}", "--serve"]
    }
  }
}
```

Or use the Cursor global config at `~/.cursor/mcp.json`.

### 3.3 OpenCode

Add to `.opencode/mcp.json` in your project root:

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/absolute/path/to/project", "--serve"]
    }
  }
}
```

### 3.4 Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS)
or `~/.config/Claude/claude_desktop_config.json` (Linux):

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/absolute/path/to/project", "--serve"]
    }
  }
}
```

### 3.5 VS Code (with MCP extension)

Add to `.vscode/mcp.json`:

```json
{
  "servers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "${workspaceFolder}", "--serve"]
    }
  }
}
```

### 3.6 Windsurf / Continue / Other Tools

Any tool that supports the MCP `stdio` transport can use CodeMap. The pattern is always:

```json
{
  "mcpServers": {
    "codemap": {
      "command": "/usr/local/bin/codemap",
      "args": ["-project", "/absolute/path/to/project", "--serve"]
    }
  }
}
```

> **Important**: Always use an **absolute path** for `-project`. Relative paths depend on
> the tool's working directory, which varies across implementations.

---

## 4. Usage Guide

Once the MCP server is configured and your tool has connected, the agent can query the project
through natural language. You don't need to teach the agent special syntax — just ask about
the project.

### 4.1 Available MCP Tools

| Tool | Description | Example Query |
|------|-------------|--------------|
| `get_project_info` | Project name, root path, module count | (no arguments) |
| `search_module` | Find module by name; empty returns all | `{"module": "order"}` |
| `related_modules` | What depends on X, and what X depends on | `{"module": "payment"}` |
| `search_route` | Find HTTP routes by path/method/module | `{"query": "/api/orders"}` |
| `search_flow` | Find data/call flows by name or trigger | `{"query": "notify"}` |
| `call_graph` | All functions a module calls | `{"module": "order"}` |
| `impact_analysis` | Who calls a given function (reverse graph) | `{"function": "CreateOrder"}` |

### 4.2 Available Resources (Markdown/JSON)

| Resource URI | Content |
|-------------|---------|
| `codemap://modules` | JSON: all modules with dependencies, types, functions, interfaces |
| `codemap://module/{name}` | JSON: single module details |
| `codemap://modules-doc/{name}` | Markdown: per-module documentation |
| `codemap://architecture/overview` | Markdown: layer stack + module registry |
| `codemap://architecture/dependencies` | Markdown: Mermaid diagram + dependency matrix |
| `codemap://routes/{name}` | Markdown: HTTP route documentation |
| `codemap://flows/{name}` | Markdown: call flow documentation |
| `codemap://callgraph/{name}` | Markdown: function call graph |

### 4.3 Example Prompts

Once CodeMap is connected, try these:

**Architecture overview:**
> Read the architecture overview and give me a 3-sentence summary of this project.

**Dependency analysis:**
> What modules depend on the database layer? Show me the dependency chain.

**Impact analysis:**
> If I modify the `payment` module, which other modules could break?

**New feature planning:**
> I want to add a notification system. Which existing modules should I touch?

**Onboarding:**
> I'm new to this project. Give me a 5-minute tour using the architecture docs.

### 4.4 Tips

- Use **Resources** for reading structured data (faster, always available)
- Use **Tools** for targeted queries (search, impact analysis)
- Run `codemap` to re-index after major refactors
- The `.codemap/` directory can be committed to Git for team sharing, or added to `.gitignore` (default)

---

## 5. Supported Languages

| Language | Status |
|----------|--------|
| Go | ✅ Full support (modules, routes, flows, call graph) |
| Java | 🔜 Planned |
| Python | 🔜 Planned |

---

## 6. Troubleshooting

### "unsupported call" on tools

Some MCP clients have issues with the `tools/call` method. CodeMap Resources (`resources/read`)
use a different transport and typically work even when tools fail.

### "another codemap server is already running"

CodeMap uses a PID lock file at `.codemap/server.lock` to prevent duplicate instances.
If the server crashed, delete the lock file manually:

```bash
rm .codemap/server.lock
```

### Index hangs or is slow

Large projects may take a few seconds. Ensure no other `codemap` process is running:

```bash
pkill codemap   # kill stale processes
codemap         # re-index
```
