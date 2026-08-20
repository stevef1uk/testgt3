package store

import "database/sql"

// Link is a saved link.
type Link struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

// InitSchema creates the links table when it does not already exist.
func InitSchema(db *sql.DB) error {
	const ddl = `
CREATE TABLE IF NOT EXISTS links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
)`
	_, err := db.Exec(ddl)
	return err
}
