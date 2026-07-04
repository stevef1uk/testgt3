package store

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func OpenDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func InitSchema(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS links (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		url TEXT NOT NULL,
		title TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(query)
	return err
}

func List() ([]Link, error) {
	rows, err := DB.Query("SELECT id, url, title, created_at FROM links ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.ID, &l.URL, &l.Title, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	return links, rows.Err()
}

func Create(url, title string) (int64, error) {
	result, err := DB.Exec("INSERT INTO links (url, title, created_at) VALUES (?, ?, ?)", url, title, time.Now())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func Delete(id int) error {
	_, err := DB.Exec("DELETE FROM links WHERE id = ?", id)
	return err
}
