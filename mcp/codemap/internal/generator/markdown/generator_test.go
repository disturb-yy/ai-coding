package markdown

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/disturb-yy/codemap/internal/model"
	"github.com/disturb-yy/codemap/internal/storage/sqlite"
)

func newTestRepo(t *testing.T) *sqlite.Repository {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sqlite.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return sqlite.NewRepository(db)
}

func TestGenerate(t *testing.T) {
	repo := newTestRepo(t)
	root := t.TempDir()

	repo.SaveModule(&model.Module{
		Name: "order", Path: "internal/order",
		Dependencies: []string{"internal/payment"},
	})
	repo.SaveModule(&model.Module{
		Name: "payment", Path: "internal/payment",
	})

	if err := Generate(repo, root); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	out := filepath.Join(root, ".codemap")

	// 检查目录结构
	for _, f := range []string{
		"INDEX.md",
		"modules/index.md",
		"modules/order.md",
		"modules/payment.md",
		"architecture/overview.md",
		"architecture/dependencies.md",
	} {
		path := filepath.Join(out, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("missing file: %s", f)
		}
	}

	// 检查模块页内容
	orderContent, _ := os.ReadFile(filepath.Join(out, "modules/order.md"))
	if !strings.Contains(string(orderContent), "internal/payment") {
		t.Error("order.md should contain dependency 'internal/payment'")
	}

	// INDEX.md 包含模块链接
	indexContent, _ := os.ReadFile(filepath.Join(out, "INDEX.md"))
	if !strings.Contains(string(indexContent), "[order](modules/order.md)") {
		t.Error("INDEX.md should link to order module")
	}
	if !strings.Contains(string(indexContent), "[payment](modules/payment.md)") {
		t.Error("INDEX.md should link to payment module")
	}
}

func TestGenerateRemovesStaleGeneratedDocs(t *testing.T) {
	repo := newTestRepo(t)
	root := t.TempDir()
	out := filepath.Join(root, ".codemap")

	if err := os.MkdirAll(filepath.Join(out, "modules"), 0755); err != nil {
		t.Fatalf("mkdir modules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(out, "modules", "stale.md"), []byte("old"), 0644); err != nil {
		t.Fatalf("write stale module: %v", err)
	}
	if err := os.WriteFile(filepath.Join(out, "codemap.db"), []byte("db"), 0644); err != nil {
		t.Fatalf("write db marker: %v", err)
	}

	if err := repo.SaveModule(&model.Module{Name: "order", Path: "internal/order"}); err != nil {
		t.Fatalf("SaveModule: %v", err)
	}
	if err := Generate(repo, root); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if _, err := os.Stat(filepath.Join(out, "modules", "stale.md")); !os.IsNotExist(err) {
		t.Fatalf("stale module doc still exists or stat failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(out, "codemap.db")); err != nil {
		t.Fatalf("codemap.db should be preserved: %v", err)
	}
}
