package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

var DB *sql.DB

var ErrNotFound = errors.New("link not found")

func List(ctx context.Context) ([]Link, error) {
	if DB == nil {
		return nil, errors.New("store database is not initialized")
	}

	rows, err := DB.QueryContext(ctx, "SELECT id, title, url, created_at FROM links ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("list links: %w", err)
	}
	defer rows.Close()

	links := make([]Link, 0)
	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.ID, &link.Title, &link.URL, &link.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan link: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate links: %w", err)
	}
	return links, nil
}

func Create(ctx context.Context, title, url string) (Link, error) {
	title = strings.TrimSpace(title)
	url = strings.TrimSpace(url)
	if title == "" {
		return Link{}, errors.New("title must not be empty")
	}
	if utf8.RuneCountInString(title) > 200 {
		return Link{}, errors.New("title must be at most 200 characters")
	}
	if url == "" {
		return Link{}, errors.New("url must not be empty")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return Link{}, errors.New("url must start with http:// or https://")
	}
	if DB == nil {
		return Link{}, errors.New("store database is not initialized")
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	result, err := DB.ExecContext(ctx,
		"INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)",
		title, url, createdAt,
	)
	if err != nil {
		return Link{}, fmt.Errorf("create link: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Link{}, fmt.Errorf("get created link ID: %w", err)
	}
	return Link{ID: id, Title: title, URL: url, CreatedAt: createdAt}, nil
}

func Delete(ctx context.Context, id int64) error {
	if DB == nil {
		return errors.New("store database is not initialized")
	}
	result, err := DB.ExecContext(ctx, "DELETE FROM links WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("count deleted links: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
