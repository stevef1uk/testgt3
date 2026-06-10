package store

import (
	"database/sql"
)

type Link struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"` // RFC3339 UTC
}

func InitSchema(db *sql.DB) error {
	const q = `CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		url TEXT NOT NULL,
		created_at TEXT NOT NULL
	);`
	_, err := db.Exec(q)
	return err
}
