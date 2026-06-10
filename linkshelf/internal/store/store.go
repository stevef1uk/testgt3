package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

var DB *sql.DB

func List(ctx context.Context) ([]Link, error) {
	const q = `SELECT id, title, url, created_at FROM links ORDER BY id DESC`
	rows, err := DB.QueryContext(ctx, q)
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
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return links, nil
}

func Create(ctx context.Context, title, url string) (Link, error) {
	// validate title
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return Link{}, fmt.Errorf("title required")
	}
	if len([]rune(trimmed)) > 200 {
		return Link{}, fmt.Errorf("title exceeds 200 characters")
	}
	// validate url
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")) {
		return Link{}, fmt.Errorf("url must start with http:// or https://")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	const q = `INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)`
	res, err := DB.ExecContext(ctx, q, trimmed, url, now)
	if err != nil {
		return Link{}, fmt.Errorf("insert link: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Link{}, fmt.Errorf("get last insert id: %w", err)
	}
	return Link{
		ID:        id,
		Title:     trimmed,
		URL:       url,
		CreatedAt: now,
	}, nil
}

func Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM links WHERE id = ?`
	res, err := DB.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("link with id %d not found", id)
	}
	return nil
}
