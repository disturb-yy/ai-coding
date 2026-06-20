package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS module (
	id      INTEGER PRIMARY KEY,
	name    TEXT,
	path    TEXT UNIQUE,
	summary TEXT
);

CREATE TABLE IF NOT EXISTS module_dependency (
	source_module TEXT,
	target_module TEXT
);

CREATE TABLE IF NOT EXISTS module_exported_type (
	source_module TEXT,
	type_name     TEXT
);

CREATE TABLE IF NOT EXISTS module_exported_func (
	source_module TEXT,
	func_name     TEXT
);

CREATE TABLE IF NOT EXISTS module_exported_method (
	source_module TEXT,
	method_name   TEXT
);

CREATE TABLE IF NOT EXISTS module_key_interface (
	source_module TEXT,
	iface_name    TEXT
);

CREATE TABLE IF NOT EXISTS route (
	id      INTEGER PRIMARY KEY,
	path    TEXT,
	method  TEXT,
	handler TEXT,
	module  TEXT
);

CREATE TABLE IF NOT EXISTS flow (
	id      TEXT PRIMARY KEY,
	name    TEXT,
	trigger TEXT
);

CREATE TABLE IF NOT EXISTS flow_step (
	flow_id  TEXT,
	step_idx INTEGER,
	step     TEXT
);

CREATE TABLE IF NOT EXISTS call_edge (
	caller_module TEXT,
	caller_func   TEXT,
	callee_module TEXT,
	callee_func   TEXT
);
`

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}

	db.SetMaxOpenConns(4)

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA synchronous=NORMAL",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("sqlite pragma %q: %w", p, err)
		}
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("sqlite schema: %w", err)
	}

	go autoCheckpoint(db, 60*time.Second)

	return db, nil
}

func autoCheckpoint(db *sql.DB, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		db.Exec("PRAGMA wal_checkpoint(PASSIVE)")
	}
}
