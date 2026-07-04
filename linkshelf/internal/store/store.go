package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// DB is the package-level database handle set at startup.
var DB *sql.DB

// List returns all links ordered by creation descending.
func List(ctx context.Context) ([]Link, error) {
	rows, err := DB.QueryContext(ctx, "SELECT id, title, url, created_at FROM links ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.ID, &l.Title, &l.URL, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	if links == nil {
		links = []Link{}
	}
	return links, rows.Err()
}

// Create inserts a new link and returns it.
func Create(ctx context.Context, title, url string) (Link, error) {
	title = strings.TrimSpace(title)
	url = strings.TrimSpace(url)

	if title == "" {
		return Link{}, fmt.Errorf("title must not be empty")
	}
	if len([]rune(title)) > 200 {
		return Link{}, fmt.Errorf("title too long (max 200 runes)")
	}
	if url == "" {
		return Link{}, fmt.Errorf("url must not be empty")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return Link{}, fmt.Errorf("url must start with http:// or https://")
	}

	res, err := DB.ExecContext(ctx, "INSERT INTO links (title, url) VALUES (?, ?)", title, url)
	if err != nil {
		return Link{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Link{}, err
	}
	var l Link
	err = DB.QueryRowContext(ctx, "SELECT id, title, url, created_at FROM links WHERE id = ?", id).
		Scan(&l.ID, &l.Title, &l.URL, &l.CreatedAt)
	return l, err
}

// Delete removes a link by ID. Returns an error if no row was affected.
func Delete(ctx context.Context, id int64) error {
	res, err := DB.ExecContext(ctx, "DELETE FROM links WHERE id = ?", id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("link not found")
	}
	return nil
}
