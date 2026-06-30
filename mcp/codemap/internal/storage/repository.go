package storage

import "github.com/disturb-yy/codemap/internal/model"

type Repository interface {
	Reset() error

	SaveModule(m *model.Module) error
	FindModule(name string) (*model.Module, error)
	SearchModule(query string) ([]*model.Module, error)

	SaveRoute(r *model.Route) error
	FindRoutes(module string) ([]*model.Route, error)

	SaveFlow(f *model.Flow) error
	FindFlows(trigger string) ([]*model.Flow, error)
	SearchFlow(query string) ([]*model.Flow, error)

	SaveCallEdge(e *model.CallEdge) error
	FindCallers(funcName string) ([]*model.CallEdge, error)
	FindCallees(module string) ([]*model.CallEdge, error)

	GetFeatureMap() ([]model.FeatureEntry, error)
	GetNavigationHints() ([]model.NavHintEntry, error)
}
