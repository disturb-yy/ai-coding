package sqlite

import (
	"database/sql"
	"fmt"

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
		"INSERT OR REPLACE INTO route(path, method, handler, module) VALUES(?, ?, ?, ?)",
		rt.Path, rt.Method, rt.Handler, rt.Module,
	)
	if err != nil {
		return fmt.Errorf("save route: %w", err)
	}
	return nil
}

func (r *Repository) FindRoutes(query string) ([]*model.Route, error) {
	rows, err := r.db.Query(
		"SELECT path, method, handler, module FROM route WHERE path LIKE ? OR method LIKE ? OR module LIKE ?",
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
