package store

import (
	"context"
	"database/sql"
	"errors"
	neturl "net/url"
	"strings"
	"time"
	"unicode/utf8"
)

// DB is the database used by the package-level persistence API.
var DB *sql.DB

var errInvalidLink = errors.New("title and url are required")
var errTitleTooLong = errors.New("title must be 200 characters or fewer")
var errInvalidURL = errors.New("url must use http or https")

// List returns saved links in newest-first order.
func List(ctx context.Context) ([]Link, error) {
	if DB == nil {
		return nil, errors.New("store database is not initialized")
	}

	rows, err := DB.QueryContext(ctx, `
		SELECT id, title, url, created_at
		FROM links
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]Link, 0)
	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.ID, &link.Title, &link.URL, &link.CreatedAt); err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

// Create validates and persists a link.
func Create(ctx context.Context, title string, url string) (Link, error) {
	if strings.TrimSpace(title) == "" || strings.TrimSpace(url) == "" {
		return Link{}, errInvalidLink
	}
	if utf8.RuneCountInString(title) > 200 {
		return Link{}, errTitleTooLong
	}
	parsedURL, err := neturl.Parse(url)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") || parsedURL.Host == "" {
		return Link{}, errInvalidURL
	}
	if DB == nil {
		return Link{}, errors.New("store database is not initialized")
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	result, err := DB.ExecContext(ctx, `
		INSERT INTO links (title, url, created_at)
		VALUES (?, ?, ?)
	`, title, url, createdAt)
	if err != nil {
		return Link{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return Link{}, err
	}
	return Link{ID: id, Title: title, URL: url, CreatedAt: createdAt}, nil
}

// Delete removes a link by ID.
func Delete(ctx context.Context, id int64) error {
	if DB == nil {
		return errors.New("store database is not initialized")
	}

	result, err := DB.ExecContext(ctx, `DELETE FROM links WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
