package store

import (
	"database/sql"
	"fmt"
)

// Link represents a bookmark with its metadata.
type Link struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"` // RFC3339 UTC
}

// InitSchema creates the links table if it does not already exist.
// It must be called once at startup before any store operations.
func InitSchema(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		created_at TEXT NOT NULL
	)`
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}
