package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// DB is the package‑level database handle. It must be set by the caller
// (e.g. main.go after opening the SQLite file).
var DB *sql.DB

// List returns all links ordered by newest first.
// It never returns a nil slice; an empty slice is returned when no rows exist.
func List(ctx context.Context) ([]Link, error) {
	if DB == nil {
		return nil, errors.New("store DB not initialized")
	}
	const query = `SELECT id, title, url, created_at FROM links ORDER BY id DESC`
	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query links: %w", err)
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
		return nil, fmt.Errorf("iterate links: %w", err)
	}
	return links, nil
}

// Create validates the input and inserts a new link.
// On success it returns the created Link with its ID and CreatedAt populated.
func Create(ctx context.Context, title, url string) (Link, error) {
	if DB == nil {
		return Link{}, errors.New("store DB not initialized")
	}
	// Validate title.
	title = strings.TrimSpace(title)
	if title == "" {
		return Link{}, errors.New("title must be non‑empty")
	}
	if len([]rune(title)) > 200 {
		return Link{}, errors.New("title exceeds 200 characters")
	}
	// Validate URL.
	url = strings.TrimSpace(url)
	if url == "" {
		return Link{}, errors.New("url must be non‑empty")
	}
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		return Link{}, errors.New("url must start with http:// or https://")
	}

	const insert = `INSERT INTO links (title, url) VALUES (?, ?)`
	res, err := DB.ExecContext(ctx, insert, title, url)
	if err != nil {
		return Link{}, fmt.Errorf("insert link: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Link{}, fmt.Errorf("retrieve last insert id: %w", err)
	}

	// Retrieve the full row (including CreatedAt).
	const sel = `SELECT id, title, url, created_at FROM links WHERE id = ?`
	row := DB.QueryRowContext(ctx, sel, id)
	var l Link
	if err := row.Scan(&l.ID, &l.Title, &l.URL, &l.CreatedAt); err != nil {
		return Link{}, fmt.Errorf("fetch created link: %w", err)
	}
	return l, nil
}

// Delete removes the link with the given id.
// Returns an error if the id does not exist.
func Delete(ctx context.Context, id int64) error {
	if DB == nil {
		return errors.New("store DB not initialized")
	}
	const del = `DELETE FROM links WHERE id = ?`
	res, err := DB.ExecContext(ctx, del, id)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	aff, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if aff == 0 {
		return fmt.Errorf("link with id %d not found", id)
	}
	return nil
}
