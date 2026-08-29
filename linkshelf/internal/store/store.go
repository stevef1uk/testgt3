package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

// DB is the package-level database handle. main.go assigns it after InitSchema.
var DB *sql.DB

// ErrNotFound is returned by Delete when the id does not exist.
var ErrNotFound = errors.New("link not found")

// List returns all links ordered by id DESC. Returns an empty slice (not nil)
// when there are no rows.
func List(ctx context.Context) ([]Link, error) {
	rows, err := DB.QueryContext(ctx, `SELECT id, title, url, created_at FROM links ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]Link, 0)
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.ID, &l.Title, &l.URL, &l.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

// Create validates title and url, inserts a new link, and returns the stored row.
func Create(ctx context.Context, title, url string) (Link, error) {
	title = strings.TrimSpace(title)
	url = strings.TrimSpace(url)

	if title == "" {
		return Link{}, errors.New("title must not be empty")
	}
	if utf8.RuneCountInString(title) > 200 {
		return Link{}, errors.New("title must be at most 200 runes")
	}
	if url == "" {
		return Link{}, errors.New("url must not be empty")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return Link{}, errors.New("url must start with http:// or https://")
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	res, err := DB.ExecContext(ctx,
		`INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)`,
		title, url, createdAt)
	if err != nil {
		return Link{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Link{}, err
	}
	return Link{ID: id, Title: title, URL: url, CreatedAt: createdAt}, nil
}

// Delete removes the link with the given id. Returns ErrNotFound if no row was deleted.
func Delete(ctx context.Context, id int64) error {
	res, err := DB.ExecContext(ctx, `DELETE FROM links WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
