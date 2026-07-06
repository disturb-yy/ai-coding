package workspace

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/disturb-yy/codemap/internal/storage"
	"github.com/disturb-yy/codemap/internal/storage/sqlite"
)

type Project struct {
	Name    string `json:"name"`
	Root    string `json:"root"`
	DBPath  string `json:"db_path"`
	Indexed bool   `json:"indexed"`
}

type Registry struct {
	root     string
	projects []Project

	mu    sync.Mutex
	dbs   map[string]*sql.DB
	repos map[string]storage.Repository
}

func New(root string) (*Registry, error) {
	projects, err := Discover(root)
	if err != nil {
		return nil, err
	}
	return &Registry{
		root:     root,
		projects: projects,
		dbs:      make(map[string]*sql.DB),
		repos:    make(map[string]storage.Repository),
	}, nil
}

func Discover(root string) ([]Project, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read workspace root: %w", err)
	}

	var projects []Project
	for _, entry := range entries {
		if !entry.IsDir() || shouldSkipDir(entry.Name()) {
			continue
		}
		projectRoot := filepath.Join(root, entry.Name())
		if !hasProjectMarker(projectRoot) {
			continue
		}
		dbPath := filepath.Join(projectRoot, ".codemap", "codemap.db")
		_, err := os.Stat(dbPath)
		projects = append(projects, Project{
			Name:    entry.Name(),
			Root:    projectRoot,
			DBPath:  dbPath,
			Indexed: err == nil,
		})
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})
	return projects, nil
}

func (r *Registry) Root() string {
	return r.root
}

func (r *Registry) Projects() []Project {
	projects := make([]Project, len(r.projects))
	copy(projects, r.projects)
	return projects
}

func (r *Registry) Resolve(project, query string) (Project, storage.Repository, error) {
	p, err := r.ResolveProject(project, query)
	if err != nil {
		return Project{}, nil, err
	}
	if !p.Indexed {
		return Project{}, nil, fmt.Errorf("project %q is not indexed; run codemap -project %s first", p.Name, p.Root)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if repo, ok := r.repos[p.Name]; ok {
		return p, repo, nil
	}
	db, err := sqlite.Open(p.DBPath)
	if err != nil {
		return Project{}, nil, fmt.Errorf("open project %q db: %w", p.Name, err)
	}
	r.dbs[p.Name] = db
	repo := sqlite.NewRepository(db)
	r.repos[p.Name] = repo
	return p, repo, nil
}

func (r *Registry) ResolveProject(project, query string) (Project, error) {
	if len(r.projects) == 0 {
		return Project{}, fmt.Errorf("no child projects found under %s", r.root)
	}
	if project != "" {
		if p, ok := r.findProject(project); ok {
			return p, nil
		}
		return Project{}, fmt.Errorf("project %q not found; available projects: %s", project, strings.Join(r.projectNames(), ", "))
	}
	if inferred, ok := r.inferProject(query); ok {
		return inferred, nil
	}
	if len(r.projects) == 1 {
		return r.projects[0], nil
	}
	return Project{}, fmt.Errorf("project required; available projects: %s", strings.Join(r.projectNames(), ", "))
}

func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []string
	for name, db := range r.dbs {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("close workspace dbs: %s", strings.Join(errs, "; "))
	}
	return nil
}

func (r *Registry) findProject(project string) (Project, bool) {
	project = strings.TrimSpace(project)
	if project == "" {
		return Project{}, false
	}
	cleanProject := filepath.Clean(project)
	baseProject := filepath.Base(cleanProject)
	for _, p := range r.projects {
		if project == p.Name || cleanProject == p.Root || baseProject == p.Name {
			return p, true
		}
	}
	return Project{}, false
}

func (r *Registry) inferProject(query string) (Project, bool) {
	tokens := splitProjectTokens(query)
	for _, p := range r.projects {
		if tokens[p.Name] {
			return p, true
		}
	}
	return Project{}, false
}

func (r *Registry) projectNames() []string {
	names := make([]string, 0, len(r.projects))
	for _, p := range r.projects {
		names = append(names, p.Name)
	}
	return names
}

func hasProjectMarker(root string) bool {
	markers := []string{
		filepath.Join(".codemap", "codemap.db"),
		"go.mod",
		"pom.xml",
		"build.gradle",
		"settings.gradle",
		".git",
	}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(root, marker)); err == nil {
			return true
		}
	}
	return false
}

func shouldSkipDir(name string) bool {
	switch name {
	case ".codemap", ".git", ".idea", ".vscode", "node_modules", "vendor":
		return true
	}
	return strings.HasPrefix(name, ".")
}

func splitProjectTokens(text string) map[string]bool {
	tokens := make(map[string]bool)
	for _, token := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return r == '/' || r == '\\' || r == '.' || r == ':' || r == '-' || r == '_' || r == ' ' || r == '\t' || r == '\n'
	}) {
		if token != "" {
			tokens[token] = true
		}
	}
	return tokens
}
