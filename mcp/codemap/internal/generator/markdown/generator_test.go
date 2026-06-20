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
