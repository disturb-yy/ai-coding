package markdown

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/disturb-yy/codemap/internal/model"
	"github.com/disturb-yy/codemap/internal/storage"
)

func Generate(repo storage.Repository, root string) error {
	modules, err := repo.SearchModule("")
	if err != nil {
		return fmt.Errorf("list modules: %w", err)
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})

	out := filepath.Join(root, ".codemap")
	dirs := []string{
		filepath.Join(out, "modules"),
		filepath.Join(out, "architecture"),
		filepath.Join(out, "routes"),
		filepath.Join(out, "flows"),
		filepath.Join(out, "callgraph"),
	}
	if err := cleanGeneratedDirs(dirs); err != nil {
		return err
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}

	if err := writeIndex(out, modules); err != nil {
		return err
	}
	if err := writeModules(repo, out, modules); err != nil {
		return err
	}
	if err := writeArchitecture(repo, out, modules); err != nil {
		return err
	}
	if err := writeRoutes(repo, out); err != nil {
		return err
	}
	if err := writeFlows(repo, out); err != nil {
		return err
	}
	if err := writeCallGraph(repo, out); err != nil {
		return err
	}
	return nil
}

func cleanGeneratedDirs(dirs []string) error {
	for _, d := range dirs {
		if err := os.RemoveAll(d); err != nil {
			return fmt.Errorf("clean %s: %w", d, err)
		}
	}
	return nil
}

func writeIndex(out string, modules []*model.Module) error {
	var b strings.Builder
	b.WriteString("# Project Index\n\n")
	b.WriteString(fmt.Sprintf("## Modules (%d)\n\n", len(modules)))
	for _, m := range modules {
		fmt.Fprintf(&b, "- [%s](modules/%s.md) — `%s`", m.Name, m.Name, m.Path)
		if len(m.Dependencies) > 0 {
			fmt.Fprintf(&b, " (%d deps)", len(m.Dependencies))
		}
		b.WriteByte('\n')
	}
	b.WriteString("\n## Architecture\n\n")
	b.WriteString("- [Overview](architecture/overview.md)\n")
	b.WriteString("- [Dependencies](architecture/dependencies.md)\n")
	b.WriteString("\n## Routes & Flows\n\n")
	b.WriteString("- [Routes](routes/index.md)\n")
	b.WriteString("- [Flows](flows/index.md)\n")
	b.WriteString("- [Call Graph](callgraph/index.md)\n")
	b.WriteString("- [Impact Analysis](callgraph/impact.md)\n")
	return os.WriteFile(filepath.Join(out, "INDEX.md"), []byte(b.String()), 0644)
}

func moduleDepsFromCallGraph(repo storage.Repository, modules []*model.Module) map[string][]string {
	depMap := make(map[string][]string)
	hasCallEdges := false
	for _, m := range modules {
		edges, err := repo.FindCallees(m.Path)
		if err != nil {
			continue
		}
		if len(edges) > 0 {
			hasCallEdges = true
		}
		seen := make(map[string]bool)
		for _, e := range edges {
			if e.CalleeModule != m.Path && !seen[e.CalleeModule] {
				depMap[m.Path] = append(depMap[m.Path], e.CalleeModule)
				seen[e.CalleeModule] = true
			}
		}
	}
	if !hasCallEdges {
		for _, m := range modules {
			depMap[m.Path] = m.Dependencies
		}
	}
	return depMap
}

func writeModules(repo storage.Repository, out string, modules []*model.Module) error {
	modDir := filepath.Join(out, "modules")
	depMap := moduleDepsFromCallGraph(repo, modules)

	var idx strings.Builder
	idx.WriteString("# Modules\n\n")
	for _, m := range modules {
		fmt.Fprintf(&idx, "- [%s](%s.md) — `%s`\n", m.Name, m.Name, m.Path)
	}
	if err := os.WriteFile(filepath.Join(modDir, "index.md"), []byte(idx.String()), 0644); err != nil {
		return err
	}
	for _, m := range modules {
		var b strings.Builder
		fmt.Fprintf(&b, "# %s\n\n## Path\n\n`%s`\n\n", m.Name, m.Path)

		b.WriteString("## Dependencies\n\n")
		deps := depMap[m.Path]
		if len(deps) == 0 {
			b.WriteString("*None*\n")
		} else {
			for _, dep := range deps {
				fmt.Fprintf(&b, "- `%s`\n", dep)
			}
		}
		b.WriteByte('\n')

		if len(m.ExportedTypes) > 0 {
			b.WriteString("## Exported Types\n\n")
			for _, t := range m.ExportedTypes {
				fmt.Fprintf(&b, "- `%s`\n", t)
			}
			b.WriteByte('\n')
		}

		if len(m.ExportedFunctions) > 0 {
			b.WriteString("## Exported Functions\n\n")
			for _, f := range m.ExportedFunctions {
				fmt.Fprintf(&b, "- `%s`\n", f)
			}
			b.WriteByte('\n')
		}

		if len(m.ExportedMethods) > 0 {
			b.WriteString("## Exported Methods\n\n")
			for _, method := range m.ExportedMethods {
				fmt.Fprintf(&b, "- `%s`\n", method)
			}
			b.WriteByte('\n')
		}

		if len(m.KeyInterfaces) > 0 {
			b.WriteString("## Key Interfaces\n\n")
			for _, iface := range m.KeyInterfaces {
				fmt.Fprintf(&b, "- `%s`\n", iface)
			}
			b.WriteByte('\n')
		}

		path := filepath.Join(modDir, m.Name+".md")
		if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
			return err
		}
	}
	return nil
}

func writeArchitecture(repo storage.Repository, out string, modules []*model.Module) error {
	archDir := filepath.Join(out, "architecture")

	// overview.md
	var ov strings.Builder
	ov.WriteString("# Architecture Overview\n\n## Layer Stack\n\n```\nSource Code\n    ↓\nGo Analyzer\n    ↓\nKnowledge Model\n    ↓\nSQLite Storage\n    ↓\nMarkdown Export\n```\n\n## Module Registry\n\n")
	for _, m := range modules {
		fmt.Fprintf(&ov, "- `%s` → %s", m.Path, m.Name)
		if len(m.Dependencies) > 0 {
			fmt.Fprintf(&ov, " (depends on: %s)", strings.Join(m.Dependencies, ", "))
		}
		ov.WriteByte('\n')
	}
	if err := os.WriteFile(filepath.Join(archDir, "overview.md"), []byte(ov.String()), 0644); err != nil {
		return err
	}

	// dependencies.md — Mermaid diagram + dependency matrix table.

	// Build module-name → path mapping for resolving deps.
	nameToPath := make(map[string]string)
	pathToName := make(map[string]string)
	allNames := make(map[string]bool)
	for _, m := range modules {
		nameToPath[m.Name] = m.Path
		pathToName[m.Path] = m.Name
		allNames[m.Name] = true
	}

	// Build dep map keyed by module name (not path).
	depMap := make(map[string]map[string]bool) // name → set of dep names
	for _, m := range modules {
		depMap[m.Name] = make(map[string]bool)
		for _, depPath := range m.Dependencies {
			if depName, ok := pathToName[depPath]; ok {
				depMap[m.Name][depName] = true
			}
		}
	}

	var db strings.Builder
	db.WriteString("# Dependency Graph\n\n```mermaid\ngraph TD\n")

	// Mermaid with sorted deterministic output
	sortedNames := sortedKeys(depMap)
	for _, name := range sortedNames {
		deps := sortedSetKeys(depMap[name])
		if len(deps) == 0 {
			fmt.Fprintf(&db, "    %s\n", sanitizeMermaidID(name))
		}
		for _, dep := range deps {
			fmt.Fprintf(&db, "    %s --> %s\n",
				sanitizeMermaidID(name), sanitizeMermaidID(dep))
		}
	}
	db.WriteString("```\n\n")

	// Dependency matrix table
	db.WriteString("## Dependency Matrix\n\n")
	db.WriteString("| Module | Dependencies |\n|--------|-------------|\n")
	for _, name := range sortedNames {
		deps := sortedSetKeys(depMap[name])
		if len(deps) == 0 {
			fmt.Fprintf(&db, "| `%s` | *None* |\n", name)
		} else {
			quoted := make([]string, len(deps))
			for i, d := range deps {
				quoted[i] = "`" + d + "`"
			}
			fmt.Fprintf(&db, "| `%s` | %s |\n", name, strings.Join(quoted, ", "))
		}
	}

	return os.WriteFile(filepath.Join(archDir, "dependencies.md"), []byte(db.String()), 0644)
}

func sortedKeys(m map[string]map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedSetKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sanitizeMermaidID(name string) string {
	r := strings.NewReplacer("-", "_", ".", "_", "/", "_", " ", "_")
	return r.Replace(name)
}

func writeRoutes(repo storage.Repository, out string) error {
	routeDir := filepath.Join(out, "routes")
	routes, err := repo.FindRoutes("")
	if err != nil {
		return fmt.Errorf("find routes: %w", err)
	}
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path < routes[j].Path
	})

	var idx strings.Builder
	idx.WriteString("# Routes\n\n")
	for _, r := range routes {
		fmt.Fprintf(&idx, "- `%s %s` → `%s` (%s)\n", r.Method, r.Path, r.Handler, r.Module)
	}
	if err := os.WriteFile(filepath.Join(routeDir, "index.md"), []byte(idx.String()), 0644); err != nil {
		return err
	}
	byMod := make(map[string][]*model.Route)
	for _, r := range routes {
		byMod[r.Module] = append(byMod[r.Module], r)
	}
	for mod, rs := range byMod {
		var b strings.Builder
		fmt.Fprintf(&b, "# Routes in %s\n\n", mod)
		for _, r := range rs {
			fmt.Fprintf(&b, "- `%s %s` → %s\n", r.Method, r.Path, r.Handler)
		}
		path := filepath.Join(routeDir, filepath.Base(mod)+".md")
		if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
			return err
		}
	}
	return nil
}

func writeFlows(repo storage.Repository, out string) error {
	flowDir := filepath.Join(out, "flows")
	flows, err := repo.SearchFlow("")
	if err != nil {
		return fmt.Errorf("search flows: %w", err)
	}
	var idx strings.Builder
	idx.WriteString("# Flows\n\n")
	for _, f := range flows {
		fmt.Fprintf(&idx, "- **%s** (trigger: %s)\n", f.Name, f.Trigger)
	}
	if err := os.WriteFile(filepath.Join(flowDir, "index.md"), []byte(idx.String()), 0644); err != nil {
		return err
	}
	for _, f := range flows {
		var b strings.Builder
		fmt.Fprintf(&b, "# %s\n\n", f.Name)
		fmt.Fprintf(&b, "- **Trigger**: %s\n\n", f.Trigger)
		b.WriteString("## Steps\n\n")
		for _, step := range f.Steps {
			fmt.Fprintf(&b, "- %s\n", step)
		}
		path := filepath.Join(flowDir, f.ID+".md")
		if err := os.WriteFile(path, []byte(b.String()), 0644); err != nil {
			return err
		}
	}
	return nil
}

func writeCallGraph(repo storage.Repository, out string) error {
	cgDir := filepath.Join(out, "callgraph")

	edges, err := repo.FindCallees("")
	if err != nil {
		return fmt.Errorf("find callees: %w", err)
	}

	var idx strings.Builder
	idx.WriteString("# Call Graph\n\n")
	if len(edges) == 0 {
		idx.WriteString("*No call edges found.*\n")
	} else {
		for _, e := range edges {
			fmt.Fprintf(&idx, "- `%s.%s` → `%s.%s`\n",
				e.CallerModule, e.CallerFunc, e.CalleeModule, e.CalleeFunc)
		}
	}
	if err := os.WriteFile(filepath.Join(cgDir, "index.md"), []byte(idx.String()), 0644); err != nil {
		return err
	}

	var imp strings.Builder
	imp.WriteString("# Impact Analysis\n\nReverse call graph — who calls whom?\n\n")
	byCallee := make(map[string][]string)
	for _, e := range edges {
		key := e.CalleeModule + "/" + e.CalleeFunc
		byCallee[key] = append(byCallee[key],
			fmt.Sprintf("`%s.%s`", e.CallerModule, e.CallerFunc))
	}
	for callee, callers := range byCallee {
		fmt.Fprintf(&imp, "## %s\n\n", callee)
		for _, c := range callers {
			fmt.Fprintf(&imp, "- %s\n", c)
		}
		imp.WriteByte('\n')
	}
	return os.WriteFile(filepath.Join(cgDir, "impact.md"), []byte(imp.String()), 0644)
}
