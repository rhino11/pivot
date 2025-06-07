package internal

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// Init initializes the database
func Init() error {
	_, err := InitDB()
	return err
}

func InitDB() (*sql.DB, error) {
	// Try to get database path from config, fallback to default
	dbPath := "./pivot.db"
	if cfg, err := loadConfig(); err == nil {
		dbPath = cfg.Database
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	schema := `
	CREATE TABLE IF NOT EXISTS issues (
		github_id INTEGER PRIMARY KEY,
		number INTEGER,
		title TEXT,
		body TEXT,
		state TEXT,
		labels TEXT,
		assignees TEXT,
		created_at TEXT,
		updated_at TEXT,
		closed_at TEXT
	);`
	_, err = db.Exec(schema)
	return db, err
}
