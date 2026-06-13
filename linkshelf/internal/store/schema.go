package store

import (
	"database/sql"
	"fmt"
)

// Link represents a bookmark entry.
type Link struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"` // RFC3339 UTC timestamp
}

// InitSchema initializes the database schema.
// It creates the 'links' table if it doesn't exist.
func InitSchema(db *sql.DB) error {
	schemaSQL := `
	CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("failed to create links table: %w", err)
	}
	return nil
}
