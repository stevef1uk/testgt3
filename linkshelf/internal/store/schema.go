package store

import (
	"database/sql"
	"fmt"
)

// Link represents a link record stored in the database.
type Link struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"` // RFC3339 UTC
}

// InitSchema creates the links table if it does not already exist.
// The table schema matches the specification:
//
//	CREATE TABLE IF NOT EXISTS links (
//	    id INTEGER PRIMARY KEY AUTOINCREMENT,
//	    title TEXT NOT NULL CHECK (length(title) <= 200),
//	    url TEXT NOT NULL CHECK (url LIKE 'http://%' OR url LIKE 'https://%'),
//	    created_at TEXT NOT NULL
//	);
func InitSchema(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}
	const ddl = `
CREATE TABLE IF NOT EXISTS links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL CHECK (length(title) <= 200),
    url TEXT NOT NULL CHECK (url LIKE 'http://%' OR url LIKE 'https://%'),
    created_at TEXT NOT NULL
);`
	_, err := db.Exec(ddl)
	if err != nil {
		return fmt.Errorf("failed to init schema: %w", err)
	}
	return nil
}
