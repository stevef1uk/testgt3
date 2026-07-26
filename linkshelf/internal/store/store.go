package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DB is the package-level database handle used by CRUD functions.
var DB *sql.DB

// List returns all links ordered by id descending.
func List(ctx context.Context) ([]Link, error) {
	rows, err := DB.QueryContext(ctx, "SELECT id, title, url, created_at FROM links ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}
	defer rows.Close()

	var links []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.ID, &l.Title, &l.URL, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan link: %w", err)
		}
		links = append(links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}
	return links, nil
}

// Create inserts a new link and returns the created Link.
func Create(ctx context.Context, title, url string) (Link, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := DB.ExecContext(ctx, "INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)", title, url, now)
	if err != nil {
		return Link{}, fmt.Errorf("create link: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Link{}, fmt.Errorf("get last insert id: %w", err)
	}
	return Link{
		ID:        id,
		Title:     title,
		URL:       url,
		CreatedAt: now,
	}, nil
}

// Delete removes a link by its ID.
func Delete(ctx context.Context, id int64) error {
	res, err := DB.ExecContext(ctx, "DELETE FROM links WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("link with id %d not found", id)
	}
	return nil
}
