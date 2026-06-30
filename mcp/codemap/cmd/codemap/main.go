package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/disturb-yy/codemap/internal/analyzer"
	"github.com/disturb-yy/codemap/internal/analyzer/golang"
	"github.com/disturb-yy/codemap/internal/analyzer/java"
	"github.com/disturb-yy/codemap/internal/generator/markdown"
	"github.com/disturb-yy/codemap/internal/mcp"
	"github.com/disturb-yy/codemap/internal/storage/sqlite"
)

func main() {
	serve := flag.Bool("serve", false, "Start MCP server (reads .codemap/codemap.db)")
	project := flag.String("project", ".", "Project root directory")
	flag.Parse()

	codemapDir := filepath.Join(*project, ".codemap")

	if *serve {
		absProject, err := filepath.Abs(*project)
		if err != nil {
			log.Fatalf("resolve project path: %v", err)
		}
		if err := os.MkdirAll(filepath.Join(absProject, ".codemap"), 0755); err != nil {
			log.Fatalf("mkdir .codemap: %v", err)
		}
		lockPath := filepath.Join(absProject, ".codemap", "server.lock")
		unlock, err := acquireLock(lockPath)
		if err != nil {
			log.Fatalf("acquire lock: %v", err)
		}
		defer unlock()

		dbPath := filepath.Join(absProject, ".codemap", "codemap.db")
		db, err := sqlite.Open(dbPath)
		if err != nil {
			log.Fatalf("open db: %v (run 'codemap' first to build index)", err)
		}
		defer db.Close()

		repo := sqlite.NewRepository(db)
		projectName := filepath.Base(absProject)
		if err := mcp.Serve(repo, projectName, absProject); err != nil {
			log.Fatal(err)
		}
		return
	}

	lang := detectLanguage(*project)
	fmt.Printf("Language: %s\n", lang)
	a := newAnalyzer(lang)

	proj, err := a.Analyze(context.Background(), *project)
	if err != nil {
		log.Fatal(err)
	}

	if err := os.MkdirAll(codemapDir, 0755); err != nil {
		log.Fatalf("mkdir .codemap: %v", err)
	}

	db, err := sqlite.Open(filepath.Join(codemapDir, "codemap.db"))
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	repo := sqlite.NewRepository(db)
	if err := repo.Reset(); err != nil {
		log.Fatalf("reset index: %v", err)
	}

	for _, m := range proj.Modules {
		if err := repo.SaveModule(m); err != nil {
			log.Fatalf("save module %s: %v", m.Path, err)
		}
	}
	for _, r := range proj.Routes {
		if err := repo.SaveRoute(r); err != nil {
			log.Fatalf("save route %s: %v", r.Path, err)
		}
	}
	for _, f := range proj.Flows {
		if err := repo.SaveFlow(f); err != nil {
			log.Fatalf("save flow %s: %v", f.ID, err)
		}
	}
	for _, e := range proj.CallEdges {
		if err := repo.SaveCallEdge(e); err != nil {
			log.Fatalf("save call edge: %v", err)
		}
	}

	fmt.Printf("Project: %s (%d modules, %d routes, %d flows, %d call edges saved)\n",
		proj.Name, len(proj.Modules), len(proj.Routes), len(proj.Flows), len(proj.CallEdges))

	if err := markdown.Generate(repo, *project); err != nil {
		log.Fatalf("generate markdown: %v", err)
	}

	ensureGitignore(*project)

	fmt.Printf("Docs: %s\n", filepath.Join(codemapDir, "INDEX.md"))

	for _, m := range proj.Modules {
		fmt.Printf("  %s (%s)", m.Name, m.Path)
		if len(m.Dependencies) > 0 {
			fmt.Printf(" -> %v", m.Dependencies)
		}
		fmt.Println()
	}
}

// detectLanguage returns the project language based on marker files/directories.
func detectLanguage(root string) string {
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
		return "go"
	}
	for _, f := range []string{"pom.xml", "build.gradle", "settings.gradle"} {
		if _, err := os.Stat(filepath.Join(root, f)); err == nil {
			return "java"
		}
	}
	for _, src := range []string{"src/main/java", "src/main/kotlin", "src"} {
		if info, err := os.Stat(filepath.Join(root, src)); err == nil && info.IsDir() {
			return "java"
		}
	}
	return "go"
}

func newAnalyzer(lang string) analyzer.Analyzer {
	switch lang {
	case "java":
		return java.New()
	default:
		return golang.New()
	}
}

func acquireLock(path string) (func(), error) {
	data, err := os.ReadFile(path)
	if err == nil {
		pidStr := strings.TrimSpace(string(data))
		if pid, convErr := strconv.Atoi(pidStr); convErr == nil && pid > 0 {
			proc, procErr := os.FindProcess(pid)
			if procErr == nil {
				if sigErr := proc.Signal(syscall.Signal(0)); sigErr == nil {
					return nil, fmt.Errorf("another codemap server is already running (PID %d); if this is wrong, remove %s", pid, path)
				}
			}
		}
		os.Remove(path)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("create lock file: %w", err)
	}
	fmt.Fprintf(f, "%d", syscall.Getpid())
	f.Close()
	return func() { os.Remove(path) }, nil
}

func ensureGitignore(root string) {
	path := filepath.Join(root, ".gitignore")
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() == ".codemap/" {
			return
		}
	}
	if _, err := f.WriteString(".codemap/\n"); err == nil {
		fmt.Println("gitignore: added .codemap/")
	}
}
