package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

var DB *sql.DB

// List returns all links ordered by creation time descending.
func List(ctx context.Context) ([]Link, error) {
	rows, err := DB.QueryContext(ctx, "SELECT id, title, url, created_at FROM links ORDER BY created_at DESC")
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
		links = []Link{}
	}
	return links, nil
}

// validateLink checks title and url constraints.
func validateLink(title, url string) error {
	if len([]rune(title)) > 200 {
		return fmt.Errorf("title must be at most 200 characters")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("url must start with http:// or https://")
	}
	return nil
}

// Create inserts a new link and returns it.
func Create(ctx context.Context, title, url string) (Link, error) {
	if err := validateLink(title, url); err != nil {
		return Link{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := DB.ExecContext(ctx,
		"INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)",
		title, url, now)
	if err != nil {
		return Link{}, fmt.Errorf("insert link: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Link{}, fmt.Errorf("last insert id: %w", err)
	}
	return Link{
		ID:        id,
		Title:     title,
		URL:       url,
		CreatedAt: now,
	}, nil
}

// Delete removes a link by id.
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
		return fmt.Errorf("link not found")
	}
	return nil
}
