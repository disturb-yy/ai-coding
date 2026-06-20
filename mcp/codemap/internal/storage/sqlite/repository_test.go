package sqlite

import (
	"path/filepath"
	"testing"

	"github.com/disturb-yy/codemap/internal/model"
)

func newTestRepo(t *testing.T) *Repository {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewRepository(db)
}

func TestSaveAndFind(t *testing.T) {
	repo := newTestRepo(t)

	m := &model.Module{
		Name:         "order",
		Path:         "internal/order",
		Dependencies: []string{"internal/payment"},
	}
	if err := repo.SaveModule(m); err != nil {
		t.Fatalf("SaveModule: %v", err)
	}

	got, err := repo.FindModule("order")
	if err != nil {
		t.Fatalf("FindModule: %v", err)
	}
	if got == nil {
		t.Fatal("expected module, got nil")
	}
	if got.Name != "order" || got.Path != "internal/order" {
		t.Errorf("got name=%q path=%q", got.Name, got.Path)
	}
	if len(got.Dependencies) != 1 || got.Dependencies[0] != "internal/payment" {
		t.Errorf("got deps=%v", got.Dependencies)
	}
}

func TestFindModule_NotFound(t *testing.T) {
	repo := newTestRepo(t)

	got, err := repo.FindModule("nonexistent")
	if err != nil {
		t.Fatalf("FindModule: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestSearchModule(t *testing.T) {
	repo := newTestRepo(t)

	repo.SaveModule(&model.Module{Name: "order", Path: "internal/order"})
	repo.SaveModule(&model.Module{Name: "payment", Path: "internal/payment"})
	repo.SaveModule(&model.Module{Name: "inventory", Path: "internal/inventory"})

	results, err := repo.SearchModule("ord")
	if err != nil {
		t.Fatalf("SearchModule: %v", err)
	}
	if len(results) != 1 || results[0].Name != "order" {
		t.Errorf("expected [order], got %d results", len(results))
	}
}

func TestSaveModule_DuplicatePath(t *testing.T) {
	repo := newTestRepo(t)

	m1 := &model.Module{Name: "order", Path: "internal/order", Dependencies: []string{"a"}}
	m2 := &model.Module{Name: "order_v2", Path: "internal/order", Dependencies: []string{"b"}}

	repo.SaveModule(m1)
	repo.SaveModule(m2)

	got, _ := repo.FindModule("order_v2")
	if got == nil {
		t.Fatal("expected module after overwrite")
	}
	if got.Name != "order_v2" {
		t.Errorf("name = %q, want order_v2", got.Name)
	}
	if len(got.Dependencies) != 1 || got.Dependencies[0] != "b" {
		t.Errorf("deps = %v, want [b]", got.Dependencies)
	}
}

func TestSaveModule_EmptyDependencies(t *testing.T) {
	repo := newTestRepo(t)

	m := &model.Module{Name: "standalone", Path: "internal/standalone"}
	repo.SaveModule(m)

	got, _ := repo.FindModule("standalone")
	if got == nil {
		t.Fatal("expected module")
	}
	if len(got.Dependencies) != 0 {
		t.Errorf("expected empty deps, got %v", got.Dependencies)
	}
}
