package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscover(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "auth", "go.mod"), "module example.com/auth\n")
	writeFile(t, filepath.Join(root, "auth", ".codemap", "codemap.db"), "db")
	writeFile(t, filepath.Join(root, "login", "go.mod"), "module example.com/login\n")
	writeFile(t, filepath.Join(root, "notes", "README.md"), "notes")

	projects, err := Discover(root)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("projects = %+v, want 2", projects)
	}
	if projects[0].Name != "auth" || !projects[0].Indexed {
		t.Fatalf("first project = %+v, want indexed auth", projects[0])
	}
	if projects[1].Name != "login" || projects[1].Indexed {
		t.Fatalf("second project = %+v, want unindexed login", projects[1])
	}
}

func TestRegistryResolveProject(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "auth", "go.mod"), "module example.com/auth\n")
	writeFile(t, filepath.Join(root, "auth", ".codemap", "codemap.db"), "db")
	writeFile(t, filepath.Join(root, "login", "go.mod"), "module example.com/login\n")
	writeFile(t, filepath.Join(root, "login", ".codemap", "codemap.db"), "db")

	registry, err := New(root)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	project, err := registry.ResolveProject("", "inspect auth login route")
	if err != nil {
		t.Fatalf("ResolveProject inferred: %v", err)
	}
	if project.Name != "auth" {
		t.Fatalf("inferred project = %q, want auth", project.Name)
	}

	project, err = registry.ResolveProject("login", "")
	if err != nil {
		t.Fatalf("ResolveProject explicit: %v", err)
	}
	if project.Name != "login" {
		t.Fatalf("explicit project = %q, want login", project.Name)
	}

	if _, err := registry.ResolveProject("", "inspect profile route"); err == nil {
		t.Fatal("expected project required error")
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
