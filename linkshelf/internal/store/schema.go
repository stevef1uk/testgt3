package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Link is a bookmark persisted by the link store.
type Link struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// InitSchema creates the persistent schema used by the link store.
func InitSchema(db *sql.DB) error {
	if db == nil {
		return fmt.Errorf("initialize schema: database is nil")
	}

	const schema = `
CREATE TABLE IF NOT EXISTS links (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	url TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL
);
`
	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("initialize schema: %w", err)
	}
	return nil
}
