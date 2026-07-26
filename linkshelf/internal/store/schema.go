package store

import (
	"database/sql"
)

// Link represents a hyperlink entry.
type Link struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"` // RFC3339 UTC
}

// InitSchema creates the links table if it does not exist.
func InitSchema(db *sql.DB) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS links (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	url TEXT NOT NULL,
	created_at TEXT NOT NULL
);
`
	_, err := db.Exec(ddl)
	return err
}
