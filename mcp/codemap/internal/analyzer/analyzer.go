package analyzer

import (
	"context"

	"github.com/disturb-yy/codemap/internal/model"
)

type Analyzer interface {
	Analyze(ctx context.Context, root string) (*model.Project, error)
}
