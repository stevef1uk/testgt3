package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// DB is the package-level database handle used by List, Create, Delete.
// It must be assigned before any CRUD operation is called.
var DB *sql.DB

// List retrieves all links ordered by id.
func List(ctx context.Context) ([]Link, error) {
	if DB == nil {
		return nil, fmt.Errorf("store.DB is nil")
	}
	rows, err := DB.QueryContext(ctx, `SELECT id, title, url, created_at FROM links ORDER BY id`)
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
	if links == nil {
		return []Link{}, nil
	}
	return links, nil
}

// Create inserts a new link and returns the created Link.
// title and url must be non-empty after trimming whitespace.
func Create(ctx context.Context, title, url string) (Link, error) {
	if DB == nil {
		return Link{}, fmt.Errorf("store.DB is nil")
	}
	title = strings.TrimSpace(title)
	url = strings.TrimSpace(url)
	if title == "" || url == "" {
		return Link{}, fmt.Errorf("title and url must not be empty")
	}

	now := time.Now().UTC().Format(time.RFC3339)

	res, err := DB.ExecContext(ctx, `INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)`, title, url, now)
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

// Delete removes a link by id. It does not error if the id does not exist.
func Delete(ctx context.Context, id int64) error {
	if DB == nil {
		return fmt.Errorf("store.DB is nil")
	}
	_, err := DB.ExecContext(ctx, `DELETE FROM links WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	return nil
}
