package store

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

// DB is the database used by the package-level store API.
var DB *sql.DB

// List returns all links ordered newest first.
func List(ctx context.Context) ([]Link, error) {
	if DB == nil {
		return []Link{}, fmt.Errorf("list links: database is nil")
	}

	rows, err := DB.QueryContext(ctx,
		`SELECT id, title, url, description, created_at FROM links ORDER BY id DESC`)
	if err != nil {
		return []Link{}, fmt.Errorf("list links: %w", err)
	}
	defer rows.Close()

	links := []Link{}
	for rows.Next() {
		var link Link
		var created string
		if err := rows.Scan(&link.ID, &link.Title, &link.URL, &link.Description, &created); err != nil {
			return []Link{}, fmt.Errorf("list links row: %w", err)
		}
		link.CreatedAt, err = time.Parse(time.RFC3339, created)
		if err != nil {
			return []Link{}, fmt.Errorf("list links timestamp: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return []Link{}, fmt.Errorf("list links rows: %w", err)
	}
	return links, nil
}

// Create validates and persists a link.
func Create(ctx context.Context, title, rawURL string) (Link, error) {
	if DB == nil {
		return Link{}, fmt.Errorf("create link: database is nil")
	}
	title = strings.TrimSpace(title)
	rawURL = strings.TrimSpace(rawURL)
	if title == "" {
		return Link{}, fmt.Errorf("create link: title is required")
	}
	if utf8.RuneCountInString(title) > 200 {
		return Link{}, fmt.Errorf("create link: title exceeds 200 characters")
	}
	if rawURL == "" {
		return Link{}, fmt.Errorf("create link: url is required")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return Link{}, fmt.Errorf("create link: url must use http or https")
	}

	now := time.Now().UTC().Truncate(time.Second)
	created := now.Format(time.RFC3339)
	result, err := DB.ExecContext(ctx,
		`INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)`,
		title, rawURL, created)
	if err != nil {
		return Link{}, fmt.Errorf("create link: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Link{}, fmt.Errorf("create link id: %w", err)
	}
	return Link{ID: id, Title: title, URL: rawURL, CreatedAt: now}, nil
}

// Delete removes a link by ID.
func Delete(ctx context.Context, id int64) error {
	if DB == nil {
		return fmt.Errorf("delete link: database is nil")
	}
	if id <= 0 {
		return fmt.Errorf("delete link: invalid id %d", id)
	}
	result, err := DB.ExecContext(ctx, `DELETE FROM links WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete link: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete link rows affected: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("delete link: id %d not found", id)
	}
	return nil
}
