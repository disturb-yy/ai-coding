package cognitive

import (
	"context"
	"strings"
	"testing"

	"github.com/disturb-yy/codemap/internal/model"
)

type fakeRepo struct {
	modules  []*model.Module
	routes   []*model.Route
	flows    []*model.Flow
	features []model.FeatureEntry
	hints    []model.NavHintEntry
	callees  []*model.CallEdge
	callers  []*model.CallEdge
}

func (r *fakeRepo) SearchModule(query string) ([]*model.Module, error) {
	if query == "" {
		return r.modules, nil
	}
	var result []*model.Module
	for _, module := range r.modules {
		if containsText(module.Name, query) || containsText(module.Path, query) {
			result = append(result, module)
		}
	}
	return result, nil
}

func (r *fakeRepo) FindRoutes(query string) ([]*model.Route, error) {
	var result []*model.Route
	for _, route := range r.routes {
		if containsText(route.Path, query) || containsText(route.Module, query) || containsText(route.Handler, query) {
			result = append(result, route)
		}
	}
	return result, nil
}

func (r *fakeRepo) SearchFlow(query string) ([]*model.Flow, error) {
	var result []*model.Flow
	for _, flow := range r.flows {
		if containsText(flow.Name, query) || containsText(flow.Trigger, query) {
			result = append(result, flow)
		}
	}
	return result, nil
}

func (r *fakeRepo) FindCallers(funcName string) ([]*model.CallEdge, error) {
	var result []*model.CallEdge
	for _, edge := range r.callers {
		if containsText(edge.CalleeFunc, funcName) || containsText(edge.CalleeModule, funcName) {
			result = append(result, edge)
		}
	}
	return result, nil
}

func (r *fakeRepo) FindCallees(module string) ([]*model.CallEdge, error) {
	var result []*model.CallEdge
	for _, edge := range r.callees {
		if containsText(edge.CallerModule, module) || containsText(edge.CallerFunc, module) {
			result = append(result, edge)
		}
	}
	return result, nil
}

func (r *fakeRepo) GetFeatureMap() ([]model.FeatureEntry, error) {
	return r.features, nil
}

func (r *fakeRepo) GetNavigationHints() ([]model.NavHintEntry, error) {
	return r.hints, nil
}

func TestFindChangePointsRanksOrderCandidates(t *testing.T) {
	repo := &fakeRepo{
		modules: []*model.Module{
			{Name: "order", Path: "internal/order", Dependencies: []string{"internal/inventory", "internal/payment"}},
			{Name: "inventory", Path: "internal/inventory"},
			{Name: "payment", Path: "internal/payment"},
		},
		routes: []*model.Route{
			{Method: "POST", Path: "/orders", Handler: "internal/order/handler.go", Module: "order"},
		},
		flows: []*model.Flow{
			{ID: "create_order", Name: "create_order_flow", Trigger: "order", Steps: []string{"reserve inventory", "charge payment"}},
		},
		features: []model.FeatureEntry{
			{Name: "Create Order", Description: "Create orders and reserve inventory", Modules: []string{"order", "inventory", "payment"}, Routes: []string{"POST /orders"}, Flows: []string{"create_order_flow"}},
		},
		hints: []model.NavHintEntry{
			{Feature: "Create Order", StartFiles: []string{"internal/order/handler.go", "internal/order/service.go"}, Routes: []string{"POST /orders"}, RelatedModules: []string{"order", "inventory", "payment"}, RelatedFlows: []string{"create_order_flow"}, Risk: []string{"order"}},
		},
		callees: []*model.CallEdge{
			{CallerModule: "order", CallerFunc: "Cancel", CalleeModule: "inventory", CalleeFunc: "Rollback"},
			{CallerModule: "order", CallerFunc: "Cancel", CalleeModule: "payment", CalleeFunc: "Refund"},
		},
	}

	service := NewService(repo)
	got, err := service.FindChangePoints(context.Background(), FindChangePointsRequest{
		Requirement: "Add order cancellation functionality",
		TopK:        3,
	})
	if err != nil {
		t.Fatalf("FindChangePoints: %v", err)
	}

	if len(got.MatchedFeatures) == 0 || got.MatchedFeatures[0].Name != "Create Order" {
		t.Fatalf("matched features = %+v", got.MatchedFeatures)
	}
	if len(got.CandidateModules) == 0 || got.CandidateModules[0].Module != "order" {
		t.Fatalf("candidate modules = %+v", got.CandidateModules)
	}
	if len(got.CandidateFiles) == 0 || got.CandidateFiles[0].File == "" {
		t.Fatalf("candidate files = %+v", got.CandidateFiles)
	}
	if len(got.RiskAreas) == 0 {
		t.Fatal("expected risk areas")
	}
	if len(got.NextActions) == 0 {
		t.Fatal("expected next actions")
	}
}

func TestTokenizeRequirementNormalizesSynonyms(t *testing.T) {
	got := TokenizeRequirement("creating cancelled orders")
	want := []string{"create", "cancel", "order"}
	if len(got) != len(want) {
		t.Fatalf("tokens = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tokens = %v, want %v", got, want)
		}
	}
}

func containsText(value, query string) bool {
	return query == "" || strings.Contains(strings.ToLower(value), strings.ToLower(query))
}
