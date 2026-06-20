# CodeMap v1 Design Specification

## 1. Project Positioning

### What is CodeMap

CodeMap is an Agent-Oriented Project Knowledge Layer.

Its goal is:

* Help AI Agents understand large codebases quickly
* Provide structured project knowledge
* Support MCP search and retrieval
* Support future multi-language analysis

CodeMap is NOT:

* Context Mode
* Code Editor
* IDE
* Call Graph Visualization Tool

---

## 2. Relationship with Other Systems

### Context Mode

Context Mode stores:

* what agent changed
* what agent decided
* what agent executed

Example:

```text
Modified order/service.go

Executed go test

Fixed TestCreateOrder
```

Context Mode = Session Memory

---

### CodeMap

CodeMap stores:

* architecture
* modules
* dependencies
* flows

Example:

```text
OrderHandler
    ↓
OrderService
    ↓
OrderRepository
```

CodeMap = Project Knowledge

---

### CodeGraph

CodeGraph stores:

```text
Symbol
Edge
Relation
```

Example:

```text
OrderService
    calls
PaymentService
```

CodeGraph = Internal Data Layer

---

## 3. Architecture

```text
Source Code
    ↓

Analyzer Layer
    ↓

Knowledge Model
    ↓

SQLite
    ↓

Markdown Export
    ↓

MCP Search
```

---

## 4. MVP Scope

### Included

* Go Language
* Module Analysis
* Import Dependency Analysis
* SQLite Storage
* Markdown Export
* MCP Search

### Excluded

* Java
* Python
* Context Memory
* Vector Search

---

## 5. Project Structure

```text
codemap/

├── cmd/
│   └── codemap/
│
├── internal/
│
│   ├── analyzer/
│   │   └── golang/
│   │
│   ├── model/
│   │
│   ├── storage/
│   │   └── sqlite/
│   │
│   ├── generator/
│   │   └── markdown/
│   │
│   ├── service/
│   │
│   └── mcp/
│
├── templates/
│
└── examples/
```

---

## 6. Core Domain Model

### Project

```go
type Project struct {
    Name string
    Root string

    Modules []*Module
}
```

### Module

```go
type Module struct {
    Name string

    Path string

    Dependencies []string
}
```

Future:

```go
type Flow struct {
    ID string
    Name string
    Trigger string

    Steps []string
}
```

---

## 7. Analyzer Layer

### Interface

```go
type Analyzer interface {
    Analyze(
        ctx context.Context,
        root string,
    ) (*model.Project,error)
}
```

---

### Go Analyzer Responsibilities

Current:

* Scan Go files
* Parse package
* Parse imports
* Build modules
* Build dependencies

Future:

* Route Analysis
* Flow Analysis
* Call Graph

---

## 8. Go Analyzer Design

### File Layout

```text
internal/analyzer/golang/

├── analyzer.go
├── imports.go
├── dependency.go
└── path.go
```

---

### analyzer.go

Responsibilities:

* Walk filesystem
* Build project
* Coordinate analysis

Should NOT:

* contain import parsing logic
* contain dependency filtering logic

Those belong in dedicated files.

---

### imports.go

Responsibilities:

```go
func ExtractImports(
    file *ast.File,
) []string
```

---

### dependency.go

Responsibilities:

```go
func IsInternalImport(
    importPath string,
) bool
```

```go
func AddDependency(
    module *model.Module,
    dependency string,
)
```

---

### path.go

Responsibilities:

```go
func ResolveModulePath(
    root string,
    file string,
) string
```

Example:

```text
/root/demo/internal/order/service.go

↓

internal/order
```

---

## 9. Dependency Rules

Only collect project internal imports.

Ignore:

```text
fmt
context
errors
gin
gorm
```

Keep:

```text
internal/order
internal/payment
internal/user
```

---

### Normalization

Convert:

```text
github.com/company/demo/internal/payment
```

To:

```text
internal/payment
```

Reason:

Module paths and import paths must match.

---

## 10. Storage Layer

SQLite is primary storage.

Markdown is generated view.

---

### Database

File:

```text
.codemap.db
```

---

### module

```sql
CREATE TABLE module(
    id INTEGER PRIMARY KEY,
    name TEXT,
    path TEXT,
    summary TEXT
);
```

---

### module_dependency

```sql
CREATE TABLE module_dependency(
    source_module TEXT,
    target_module TEXT
);
```

---

## 11. Repository Layer

Interface:

```go
type Repository interface {

    SaveModule(
        *model.Module,
    ) error

    FindModule(
        string,
    ) (*model.Module,error)

    SearchModule(
        string,
    ) ([]*model.Module,error)
}
```

---

## 12. Markdown Export

IMPORTANT:

Markdown is NOT source of truth.

SQLite is source of truth.

---

Directory Layout

```text
.codemap/

├── INDEX.md
│
├── modules/
│   ├── index.md
│   ├── order.md
│   └── user.md
│
└── architecture/
    ├── overview.md
    └── dependencies.md
```

---

### INDEX.md

Maximum size:

```text
5KB
```

Purpose:

Project Entry

---

### Module Page

Example:

```markdown
# order

Path

internal/order

Dependencies

- payment
- inventory
```

Maximum size:

```text
10KB
```

---

## 13. MCP Design

Uses [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) v1.6.1 — the official MCP Go SDK — for JSON-RPC transport and tool registration.

### Search Module

```json
{
  "module":"order"
}
```

Response:

```json
{
  "path":"internal/order",
  "dependencies":[
    "payment",
    "inventory"
  ]
}
```

---

### Related Modules

```json
{
  "module":"order"
}
```

Response:

```json
{
  "dependencies":[
    "payment",
    "inventory"
  ]
}
```

---

## 14. Coding Standards

### Formatting

Mandatory:

```bash
gofmt -w .
goimports -w .
```

---

### Function Calls

Short calls should remain on one line.

Preferred:

```go
return a.analyzeFile(root, path, moduleIndex)
```

Not:

```go
return a.analyzeFile(
    root,
    path,
    moduleIndex,
)
```

Unless line length exceeds project limit.

---

### Struct Initialization

Preferred:

```go
module := &model.Module{
    Name: file.Name.Name,
    Path: modulePath,
}
```

---

### Comments

All exported symbols must contain GoDoc comments.

Example:

```go
// ExtractImports extracts import paths from a Go file.
func ExtractImports(
    file *ast.File,
) []string
```

---

## 15. Development Roadmap

### Phase 1

Project Model

Go Analyzer

Dependency Analysis

JSON Export

---

### Phase 2

SQLite

Repository

Search

---

### Phase 3

Markdown Export

---

### Phase 4

MCP

---

### Phase 5

Route Analysis

Flow Analysis

---

### Phase 6

Java Support

Python Support

````

---

## Success Criteria

Agent can answer:

```text
Where is Order module?

Which modules depend on Payment?

What modules are related to Order?

How is the project structured?
````

without scanning the entire repository.
