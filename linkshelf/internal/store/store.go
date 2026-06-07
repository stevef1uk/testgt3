package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// DB is the global database handle used by the store package.
// It is expected to be set by the application entrypoint after opening a connection.
var DB *sql.DB

// List returns all links ordered by ID.
// It returns a slice of Link and any error encountered while querying.
func List(ctx context.Context) ([]Link, error) {
	if DB == nil {
		return nil, errors.New("store DB not initialized")
	}
	const query = `SELECT id, title, url, created_at FROM links ORDER BY id`
	rows, err := DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query links: %w", err)
	}
	defer rows.Close()

	var result []Link
	for rows.Next() {
		var l Link
		if err := rows.Scan(&l.ID, &l.Title, &l.URL, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan link row: %w", err)
		}
		result = append(result, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return result, nil
}

// Create validates the input and inserts a new link record.
// It returns the created Link with its generated ID and timestamp.
func Create(ctx context.Context, title, url string) (Link, error) {
	if DB == nil {
		return Link{}, errors.New("store DB not initialized")
	}
	// Validation per architecture.
	if strings.TrimSpace(title) == "" {
		return Link{}, errors.New("title must be non‑empty")
	}
	if len([]rune(title)) > 200 {
		return Link{}, errors.New("title exceeds 200 characters")
	}
	if strings.TrimSpace(url) == "" {
		return Link{}, errors.New("url must be non‑empty")
	}
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		return Link{}, errors.New("url must start with http:// or https://")
	}

	const stmt = `INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)`
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := DB.ExecContext(ctx, stmt, title, url, now)
	if err != nil {
		return Link{}, fmt.Errorf("insert link: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Link{}, fmt.Errorf("retrieve insert id: %w", err)
	}
	return Link{
		ID:        id,
		Title:     title,
		URL:       url,
		CreatedAt: now,
	}, nil
}

// Delete removes the link with the given ID.
// It returns an error if the operation fails.
func Delete(ctx context.Context, id int64) error {
	if DB == nil {
		return errors.New("store DB not initialized")
	}
	const stmt = `DELETE FROM links WHERE id = ?`
	_, err := DB.ExecContext(ctx, stmt, id)
	if err != nil {
		return fmt.Errorf("delete link id %d: %w", id, err)
	}
	return nil
}
