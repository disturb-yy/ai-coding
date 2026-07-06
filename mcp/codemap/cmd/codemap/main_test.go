package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProjectRoot_FromSubdir(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	subdir := filepath.Join(root, "internal", "api")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}

	got, err := resolveProjectRoot(subdir)
	if err != nil {
		t.Fatalf("resolveProjectRoot: %v", err)
	}
	if got != root {
		t.Fatalf("resolveProjectRoot(%q) = %q, want %q", subdir, got, root)
	}
}

func TestResolveProjectRoot_FromCodemapDir(t *testing.T) {
	root := t.TempDir()
	codemapDir := filepath.Join(root, ".codemap")
	if err := os.MkdirAll(codemapDir, 0755); err != nil {
		t.Fatalf("mkdir .codemap: %v", err)
	}
	if err := os.WriteFile(filepath.Join(codemapDir, "codemap.db"), []byte("db"), 0644); err != nil {
		t.Fatalf("write codemap.db: %v", err)
	}
	subdir := filepath.Join(root, "api")
	if err := os.MkdirAll(subdir, 0755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}

	got, err := resolveProjectRoot(subdir)
	if err != nil {
		t.Fatalf("resolveProjectRoot: %v", err)
	}
	if got != root {
		t.Fatalf("resolveProjectRoot(%q) = %q, want %q", subdir, got, root)
	}
}

func TestResolveProjectRootForMode_WorkspaceDoesNotClimb(t *testing.T) {
	parent := t.TempDir()
	if err := os.Mkdir(filepath.Join(parent, ".git"), 0755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	workspaceRoot := filepath.Join(parent, "projects")
	if err := os.Mkdir(workspaceRoot, 0755); err != nil {
		t.Fatalf("mkdir workspace: %v", err)
	}

	got, err := resolveProjectRootForMode(workspaceRoot, true)
	if err != nil {
		t.Fatalf("resolveProjectRootForMode: %v", err)
	}
	if got != workspaceRoot {
		t.Fatalf("resolveProjectRootForMode(%q, true) = %q, want %q", workspaceRoot, got, workspaceRoot)
	}
}
