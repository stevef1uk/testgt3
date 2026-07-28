package store

import (
	"database/sql"
	"fmt"
)

// Link represents a stored hyperlink.
// CreatedAt is stored as RFC3339 UTC string.
type Link struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"` // RFC3339 UTC
}

// InitSchema ensures the SQLite schema exists.
// It creates the `links` table if it does not already exist.
func InitSchema(db *sql.DB) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TEXT NOT NULL
);`
	_, err := db.Exec(ddl)
	if err != nil {
		return fmt.Errorf("failed to init schema: %w", err)
	}
	return nil
}
