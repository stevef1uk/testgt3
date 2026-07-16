package store

import (
	"database/sql"
)

// DB is the package‑level database handle. It will be set by the application
// entrypoint (e.g., main.go) before any store functions are called.
var DB *sql.DB

// List returns all stored links ordered by their ID.
func List() ([]Link, error) {
	rows, err := DB.Query(`SELECT id, title, url, created_at FROM links ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := []Link{}
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.ID, &l.Title, &l.URL, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

// Create inserts a new Link into the database and returns its generated ID.
func Create(l Link) (int64, error) {
	res, err := DB.Exec(
		`INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)`,
		l.Title, l.URL, l.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Delete removes the link with the specified ID from the database.
func Delete(id int64) error {
	_, err := DB.Exec(`DELETE FROM links WHERE id = ?`, id)
	return err
}
