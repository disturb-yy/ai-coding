package sqlite

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/disturb-yy/codemap/internal/model"
	"github.com/disturb-yy/codemap/internal/storage"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) SaveModule(m *model.Module) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("save module begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT OR REPLACE INTO module(name, path, summary) VALUES(?, ?, ?)",
		m.Name, m.Path, "",
	)
	if err != nil {
		return fmt.Errorf("save module insert: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM module_dependency WHERE source_module = ?", m.Path); err != nil {
		return fmt.Errorf("save module delete deps: %w", err)
	}
	for _, dep := range m.Dependencies {
		if _, err := tx.Exec(
			"INSERT INTO module_dependency(source_module, target_module) VALUES(?, ?)",
			m.Path, dep,
		); err != nil {
			return fmt.Errorf("save module insert dep: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM module_exported_type WHERE source_module = ?", m.Path); err != nil {
		return fmt.Errorf("save module delete types: %w", err)
	}
	for _, t := range m.ExportedTypes {
		if _, err := tx.Exec(
			"INSERT INTO module_exported_type(source_module, type_name) VALUES(?, ?)",
			m.Path, t,
		); err != nil {
			return fmt.Errorf("save module insert type: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM module_exported_func WHERE source_module = ?", m.Path); err != nil {
		return fmt.Errorf("save module delete funcs: %w", err)
	}
	for _, f := range m.ExportedFunctions {
		if _, err := tx.Exec(
			"INSERT INTO module_exported_func(source_module, func_name) VALUES(?, ?)",
			m.Path, f,
		); err != nil {
			return fmt.Errorf("save module insert func: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM module_exported_method WHERE source_module = ?", m.Path); err != nil {
		return fmt.Errorf("save module delete methods: %w", err)
	}
	for _, method := range m.ExportedMethods {
		if _, err := tx.Exec(
			"INSERT INTO module_exported_method(source_module, method_name) VALUES(?, ?)",
			m.Path, method,
		); err != nil {
			return fmt.Errorf("save module insert method: %w", err)
		}
	}

	if _, err := tx.Exec("DELETE FROM module_key_interface WHERE source_module = ?", m.Path); err != nil {
		return fmt.Errorf("save module delete ifaces: %w", err)
	}
	for _, iface := range m.KeyInterfaces {
		if _, err := tx.Exec(
			"INSERT INTO module_key_interface(source_module, iface_name) VALUES(?, ?)",
			m.Path, iface,
		); err != nil {
			return fmt.Errorf("save module insert iface: %w", err)
		}
	}

	return tx.Commit()
}

func (r *Repository) FindModule(name string) (*model.Module, error) {
	m, err := r.findModuleByName(name)
	if err != nil {
		return nil, err
	}
	if m != nil {
		return m, nil
	}
	return r.findModuleByPath(name)
}

func (r *Repository) findModuleByName(name string) (*model.Module, error) {
	row := r.db.QueryRow("SELECT name, path FROM module WHERE name = ?", name)
	m := &model.Module{}
	err := row.Scan(&m.Name, &m.Path)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find module: %w", err)
	}
	if err := r.loadModuleExtras(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *Repository) findModuleByPath(path string) (*model.Module, error) {
	rows, err := r.db.Query("SELECT name, path FROM module WHERE path LIKE ?", "%"+path+"%")
	if err != nil {
		return nil, fmt.Errorf("find module by path: %w", err)
	}
	defer rows.Close()

	var found []*model.Module
	for rows.Next() {
		m := &model.Module{}
		if err := rows.Scan(&m.Name, &m.Path); err != nil {
			return nil, fmt.Errorf("find module by path scan: %w", err)
		}
		if err := r.loadModuleExtras(m); err != nil {
			return nil, err
		}
		found = append(found, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, nil
	}
	if len(found) == 1 {
		return found[0], nil
	}
	return nil, fmt.Errorf("ambiguous path %q matched %d modules: use exact module name", path, len(found))
}

func (r *Repository) SearchModule(query string) ([]*model.Module, error) {
	rows, err := r.db.Query("SELECT name, path FROM module WHERE name LIKE ?", "%"+query+"%")
	if err != nil {
		return nil, fmt.Errorf("search module: %w", err)
	}
	defer rows.Close()

	var modules []*model.Module
	for rows.Next() {
		m := &model.Module{}
		if err := rows.Scan(&m.Name, &m.Path); err != nil {
			return nil, fmt.Errorf("search module scan: %w", err)
		}
		if err := r.loadModuleExtras(m); err != nil {
			return nil, err
		}
		modules = append(modules, m)
	}
	return modules, rows.Err()
}

func (r *Repository) loadModuleExtras(m *model.Module) error {
	deps, err := r.loadDependencies(m.Path)
	if err != nil {
		return err
	}
	m.Dependencies = deps

	types, err := r.loadStrings("SELECT type_name FROM module_exported_type WHERE source_module = ?", m.Path)
	if err != nil {
		return err
	}
	m.ExportedTypes = types

	funcs, err := r.loadStrings("SELECT func_name FROM module_exported_func WHERE source_module = ?", m.Path)
	if err != nil {
		return err
	}
	m.ExportedFunctions = funcs

	methods, err := r.loadStrings("SELECT method_name FROM module_exported_method WHERE source_module = ?", m.Path)
	if err != nil {
		return err
	}
	m.ExportedMethods = methods

	ifaces, err := r.loadStrings("SELECT iface_name FROM module_key_interface WHERE source_module = ?", m.Path)
	if err != nil {
		return err
	}
	m.KeyInterfaces = ifaces

	return nil
}

func (r *Repository) SaveRoute(rt *model.Route) error {
	_, err := r.db.Exec(
		"INSERT INTO route(path, method, handler, module) VALUES(?, ?, ?, ?)",
		rt.Path, rt.Method, rt.Handler, rt.Module,
	)
	if err != nil {
		return fmt.Errorf("save route: %w", err)
	}
	return nil
}

func (r *Repository) FindRoutes(query string) ([]*model.Route, error) {
	rows, err := r.db.Query(
		"SELECT path, method, handler, module FROM route WHERE path LIKE ? OR module LIKE ? OR handler LIKE ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("find routes: %w", err)
	}
	defer rows.Close()

	var routes []*model.Route
	for rows.Next() {
		rt := &model.Route{}
		if err := rows.Scan(&rt.Path, &rt.Method, &rt.Handler, &rt.Module); err != nil {
			return nil, fmt.Errorf("scan route: %w", err)
		}
		routes = append(routes, rt)
	}
	return routes, rows.Err()
}

func (r *Repository) SaveFlow(f *model.Flow) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("save flow begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT OR REPLACE INTO flow(id, name, trigger) VALUES(?, ?, ?)",
		f.ID, f.Name, f.Trigger,
	)
	if err != nil {
		return fmt.Errorf("save flow insert: %w", err)
	}

	_, err = tx.Exec("DELETE FROM flow_step WHERE flow_id = ?", f.ID)
	if err != nil {
		return fmt.Errorf("save flow delete steps: %w", err)
	}
	for i, step := range f.Steps {
		_, err = tx.Exec(
			"INSERT INTO flow_step(flow_id, step_idx, step) VALUES(?, ?, ?)",
			f.ID, i, step,
		)
		if err != nil {
			return fmt.Errorf("save flow insert step: %w", err)
		}
	}
	return tx.Commit()
}

func (r *Repository) FindFlows(trigger string) ([]*model.Flow, error) {
	rows, err := r.db.Query(
		"SELECT id, name, trigger FROM flow WHERE trigger LIKE ?",
		"%"+trigger+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("find flows: %w", err)
	}
	defer rows.Close()
	return r.scanFlows(rows)
}

func (r *Repository) SearchFlow(query string) ([]*model.Flow, error) {
	rows, err := r.db.Query(
		"SELECT id, name, trigger FROM flow WHERE name LIKE ?",
		"%"+query+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("search flow: %w", err)
	}
	defer rows.Close()
	return r.scanFlows(rows)
}

func (r *Repository) scanFlows(rows *sql.Rows) ([]*model.Flow, error) {
	var flows []*model.Flow
	for rows.Next() {
		f := &model.Flow{}
		if err := rows.Scan(&f.ID, &f.Name, &f.Trigger); err != nil {
			return nil, fmt.Errorf("scan flow: %w", err)
		}
		steps, err := r.loadFlowSteps(f.ID)
		if err != nil {
			return nil, err
		}
		f.Steps = steps
		flows = append(flows, f)
	}
	return flows, rows.Err()
}

func (r *Repository) loadFlowSteps(flowID string) ([]string, error) {
	rows, err := r.db.Query(
		"SELECT step FROM flow_step WHERE flow_id = ? ORDER BY step_idx",
		flowID,
	)
	if err != nil {
		return nil, fmt.Errorf("load flow steps: %w", err)
	}
	defer rows.Close()

	var steps []string
	for rows.Next() {
		var step string
		if err := rows.Scan(&step); err != nil {
			return nil, fmt.Errorf("scan flow step: %w", err)
		}
		steps = append(steps, step)
	}
	return steps, rows.Err()
}

func (r *Repository) loadDependencies(modulePath string) ([]string, error) {
	rows, err := r.db.Query(
		"SELECT target_module FROM module_dependency WHERE source_module = ?",
		modulePath,
	)
	if err != nil {
		return nil, fmt.Errorf("load deps: %w", err)
	}
	defer rows.Close()

	var deps []string
	for rows.Next() {
		var dep string
		if err := rows.Scan(&dep); err != nil {
			return nil, fmt.Errorf("load deps scan: %w", err)
		}
		deps = append(deps, dep)
	}
	return deps, rows.Err()
}

func (r *Repository) loadStrings(query, arg string) ([]string, error) {
	rows, err := r.db.Query(query, arg)
	if err != nil {
		return nil, fmt.Errorf("load strings: %w", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, fmt.Errorf("load strings scan: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}

var _ storage.Repository = (*Repository)(nil)

func (r *Repository) SaveCallEdge(e *model.CallEdge) error {
	_, err := r.db.Exec(
		"INSERT INTO call_edge(caller_module, caller_func, callee_module, callee_func) VALUES(?, ?, ?, ?)",
		e.CallerModule, e.CallerFunc, e.CalleeModule, e.CalleeFunc,
	)
	if err != nil {
		return fmt.Errorf("save call edge: %w", err)
	}
	return nil
}

func (r *Repository) FindCallers(funcName string) ([]*model.CallEdge, error) {
	rows, err := r.db.Query(
		"SELECT caller_module, caller_func, callee_module, callee_func FROM call_edge WHERE callee_func LIKE ?",
		"%"+funcName+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("find callers: %w", err)
	}
	defer rows.Close()
	return scanCallEdges(rows)
}

func (r *Repository) FindCallees(module string) ([]*model.CallEdge, error) {
	rows, err := r.db.Query(
		"SELECT caller_module, caller_func, callee_module, callee_func FROM call_edge WHERE caller_module LIKE ?",
		"%"+module+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("find callees: %w", err)
	}
	defer rows.Close()
	return scanCallEdges(rows)
}

func scanCallEdges(rows *sql.Rows) ([]*model.CallEdge, error) {
	var edges []*model.CallEdge
	for rows.Next() {
		e := &model.CallEdge{}
		if err := rows.Scan(&e.CallerModule, &e.CallerFunc, &e.CalleeModule, &e.CalleeFunc); err != nil {
			return nil, fmt.Errorf("scan call edge: %w", err)
		}
		edges = append(edges, e)
	}
	return edges, rows.Err()
}

// GetFeatureMap derives business features from flows, routes, and modules.
func (r *Repository) GetFeatureMap() ([]model.FeatureEntry, error) {
	flowRows, err := r.db.Query("SELECT DISTINCT trigger FROM flow WHERE trigger != ''")
	if err != nil {
		return nil, fmt.Errorf("get feature map triggers: %w", err)
	}
	defer flowRows.Close()

	var triggers []string
	for flowRows.Next() {
		var t string
		if err := flowRows.Scan(&t); err != nil {
			return nil, fmt.Errorf("scan trigger: %w", err)
		}
		triggers = append(triggers, t)
	}
	if err := flowRows.Err(); err != nil {
		return nil, err
	}

	if len(triggers) == 0 {
		return r.featureMapFromModules()
	}

	var features []model.FeatureEntry
	seen := make(map[string]bool)
	for _, trigger := range triggers {
		name := flowNameToFeatureName(trigger)
		if seen[name] {
			continue
		}
		seen[name] = true

		feat := model.FeatureEntry{Name: name, Modules: []string{trigger}}

		flows, _ := r.SearchFlow(trigger)
		for _, f := range flows {
			feat.Flows = appendIfNew(feat.Flows, f.Name)
		}

		routes, _ := r.FindRoutes(trigger)
		for _, rt := range routes {
			feat.Routes = appendIfNew(feat.Routes, rt.Method+" "+rt.Path)
		}

		callees, _ := r.FindCallees(trigger)
		for _, e := range callees {
			feat.Modules = appendIfNew(feat.Modules, e.CalleeModule)
		}

		callers, _ := r.FindCallers(trigger)
		for _, e := range callers {
			feat.Modules = appendIfNew(feat.Modules, e.CallerModule)
		}

		features = append(features, feat)
	}

	return features, nil
}

func (r *Repository) featureMapFromModules() ([]model.FeatureEntry, error) {
	modules, err := r.SearchModule("")
	if err != nil {
		return nil, fmt.Errorf("feature map from modules: %w", err)
	}
	var features []model.FeatureEntry
	for _, m := range modules {
		feat := model.FeatureEntry{
			Name:    flowNameToFeatureName(m.Name),
			Modules: []string{m.Path},
		}
		callees, _ := r.FindCallees(m.Path)
		for _, e := range callees {
			feat.Modules = appendIfNew(feat.Modules, e.CalleeModule)
		}
		features = append(features, feat)
	}
	return features, nil
}

// GetNavigationHints derives navigation guidance from flows, routes, call graph, and modules.
func (r *Repository) GetNavigationHints() ([]model.NavHintEntry, error) {
	feats, err := r.GetFeatureMap()
	if err != nil {
		return nil, fmt.Errorf("get navigation hints: %w", err)
	}

	var hints []model.NavHintEntry
	for _, feat := range feats {
		hint := model.NavHintEntry{
			Feature:        feat.Name,
			Routes:         feat.Routes,
			RelatedModules: feat.Modules,
			RelatedFlows:   feat.Flows,
		}

		for _, rt := range feat.Routes {
			hint.StartFiles = appendIfNew(hint.StartFiles, routeToStartFile(rt))
		}
		if len(hint.StartFiles) == 0 {
			for _, mod := range feat.Modules {
				hint.StartFiles = appendIfNew(hint.StartFiles, mod+"/...")
			}
		}

		hint.Risk = r.identifyRisks(feat.Modules)
		hints = append(hints, hint)
	}

	return hints, nil
}

func (r *Repository) identifyRisks(modules []string) []string {
	var risks []string
	for _, mod := range modules {
		callees, err := r.FindCallees(mod)
		if err != nil {
			continue
		}
		if len(callees) >= 3 {
			risks = append(risks, mod)
		}
	}
	return risks
}

func flowNameToFeatureName(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "/", " ")
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func routeToStartFile(route string) string {
	parts := strings.Fields(route)
	if len(parts) < 2 {
		return ""
	}
	path := parts[1]
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	path = strings.ReplaceAll(path, "-", "_")
	path = strings.ReplaceAll(path, "/", "_")
	return path
}

func appendIfNew(slice []string, s string) []string {
	for _, existing := range slice {
		if existing == s {
			return slice
		}
	}
	return append(slice, s)
}
