# CodeMap — Installation & Usage Guide

CodeMap is an agent-oriented project knowledge layer. It indexes Go and Java codebases into a
SQLite database and exposes structured project knowledge (modules, dependencies, call graphs,
flows, routes, features, navigation hints) through the MCP protocol. Language is auto-detected — no flags needed.

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

The same command works for Go and Java — CodeMap detects the language automatically.

```bash
# Go project (detects go.mod)
codemap -project /path/to/go-project

# Java project (detects pom.xml / build.gradle / src/main/java)
codemap -project /path/to/java-project
```

**Language detection rules:**

| Condition | Language |
|-----------|----------|
| `go.mod` exists | Go |
| `pom.xml`, `build.gradle`, or `settings.gradle` exists | Java |
| `src/main/java` directory exists | Java |
| None of the above | Go (default) |

This produces:

```
your-project/
├── .codemap/
│   ├── codemap.db          # SQLite database (source of truth)
│   ├── INDEX.md            # Project entry point
│   ├── modules/            # Per-module Markdown docs (deps, types, methods, interfaces)
│   ├── architecture/       # Overview + dependency graph (Mermaid)
│   ├── routes/             # HTTP route docs (Go only)
│   ├── flows/              # Cross-module call flows (Go only)
│   └── callgraph/          # Function-level call graph (Go only)
```

A `.gitignore` entry for `.codemap/` is added automatically.

> Re-run `codemap` after significant code changes to keep the index up to date.

---

## 3. MCP Server Configuration

CodeMap exposes two kinds of MCP interfaces:

| Channel | Method | What It Provides |
|---------|--------|-----------------|
| **Tools** | `tools/call` | 10 query tools (search, impact analysis, call graph, list, etc.) |
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

### 4.0 Manual Testing (without an MCP client)

You can test CodeMap tools directly from the command line without configuring any IDE.

**Stop any running server first:**

```bash
pkill codemap
rm -f .codemap/server.lock
```

#### One-shot: full MCP handshake + tool call

The server speaks JSON-RPC over stdio. Send `initialize`, `notifications/initialized`,
then `tools/call`. The last line of stdout is the tool result:

```bash
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"cli","version":"1"}}}\n{"jsonrpc":"2.0","method":"notifications/initialized"}\n{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_modules"}}\n' \
  | codemap -project . --serve 2>/dev/null \
  | tail -1 | python3 -m json.tool | head -40
```

#### All 8 tools with their JSON arguments

| Tool | Arguments |
|------|-----------|
| `get_project_info` | `{}` |
| `list_modules` | `{}` |
| `search_module` | `{"module":"market"}` |
| `related_modules` | `{"module":"strategy"}` |
| `search_route` | `{"query":"/api"}` |
| `search_flow` | `{"query":"notify"}` |
| `call_graph` | `{"module":"notify"}` |
| `impact_analysis` | `{"function":"NewDispatcher"}` |

Example — call `search_module` for "market":

```bash
printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"cli","version":"1"}}}\n{"jsonrpc":"2.0","method":"notifications/initialized"}\n{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"search_module","arguments":{"module":"market"}}}\n' \
  | codemap -project . --serve 2>/dev/null \
  | tail -1
```

#### Helper script (recommended)

Save as `mcp-call.sh`:

```bash
#!/bin/bash
# Usage: ./mcp-call.sh <tool_name> [json_args]
#   ./mcp-call.sh list_modules
#   ./mcp-call.sh search_module '{"module":"market"}'
#   ./mcp-call.sh impact_analysis '{"function":"NewDispatcher"}'

TOOL=${1:?tool name required}
ARGS=${2:-"{}"}

printf '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"cli","version":"1"}}}\n{"jsonrpc":"2.0","method":"notifications/initialized"}\n{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"%s","arguments":%s}}\n' "$TOOL" "$ARGS" \
  | codemap -project . --serve 2>/dev/null \
  | tail -1 | python3 -m json.tool 2>/dev/null || tail -1
```

```bash
chmod +x mcp-call.sh
./mcp-call.sh list_modules
./mcp-call.sh search_module '{"module":"strategy"}'
./mcp-call.sh impact_analysis '{"function":"NewDispatcher"}'
```

#### Direct SQLite queries (no server needed)

```bash
# List all modules
sqlite3 .codemap/codemap.db "SELECT name, path FROM module ORDER BY path;"

# Dependency counts
sqlite3 .codemap/codemap.db \
  "SELECT source_module, COUNT(*) c FROM module_dependency GROUP BY source_module ORDER BY c DESC;"

# Callers of a function
sqlite3 .codemap/codemap.db \
  "SELECT caller_module || '.' || caller_func FROM call_edge WHERE callee_func LIKE '%Dispatch%';"
```

### 4.1 Available MCP Tools

| Tool | Description | Example Query |
|------|-------------|--------------|
| `get_project_info` | Project name, root path, module count | (no arguments) |
| `list_modules` | List all modules with full details | (no arguments) |
| `search_module` | Find module by name; empty returns all | `{"module": "market"}` |
| `related_modules` | What depends on X, and what X depends on | `{"module": "strategy"}` |
| `search_route` | Find HTTP routes by path/method/module | `{"query": "/api"}` |
| `search_flow` | Find data/call flows by name or trigger | `{"query": "notify"}` |
| `call_graph` | All functions a module calls | `{"module": "notify"}` |
| `impact_analysis` | Who calls a given function (reverse graph) | `{"function": "NewDispatcher"}` |
| `get_feature_map` | Business feature map — features, modules, routes, flows | (no arguments) |
| `get_navigation_hints` | Navigation guidance — entry files, related modules, risks | (no arguments) |

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

**Feature discovery:**
> What business features does this project have? Show me the feature map.

**Code navigation:**
> I need to work on the payment feature. Where should I start? What files are the entry points?

### 4.4 Tips

- Use **Resources** for reading structured data (faster, always available)
- Use **Tools** for targeted queries (search, impact analysis)
- Run `codemap` to re-index after major refactors
- The `.codemap/` directory can be committed to Git for team sharing, or added to `.gitignore` (default)

---

## 5. Supported Languages

| Language | Status | Capabilities |
|----------|--------|-------------|
| **Go** | ✅ Full | Modules, dependencies, exported types/funcs/methods/interfaces, HTTP routes, call flows, call graph, impact analysis |
| **Java** | ✅ Basic | Modules, dependencies (internal imports, filesystem-verified), exported types, methods (`ClassName.method`), key interfaces |
| Python | 🔜 Planned | — |

### 5.1 Go vs Java Analysis

| Dimension | Go | Java |
|-----------|----|------|
| Module detection | Directory-based (`go.mod` internal imports) | Package-based (src directory tree) |
| Dependency resolution | `go.mod` import path matching | Filesystem directory verification |
| Route extraction | AST-level HTTP handler detection | Not yet supported |
| Call flows | AST-level cross-module call chains | Not yet supported |
| Call graph | Function-level edges | Not yet supported |
| Type extraction | `ast.TypeSpec` (struct/interface) | `public class` / `public interface` / `public enum` |
| Method extraction | `Receiver.Method` | `ClassName.method` (instance + static) |
| Interface detection | `ast.InterfaceType` | `public interface` keyword |
| Annotation handling | N/A | Skips `@Override`, `@Deprecated` etc. |
| Multi-class files | One package per directory | One class per file (primary class only) |

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
