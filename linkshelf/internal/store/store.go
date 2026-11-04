package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Store provides persistence for links.
type Store struct {
	DB *sql.DB
}

// New returns a store backed by db.
func New(db *sql.DB) *Store {
	return &Store{DB: db}
}

// CreateLink persists a link and returns the stored record.
func (s *Store) CreateLink(ctx context.Context, url, title string) (Link, error) {
	if s == nil || s.DB == nil {
		return Link{}, fmt.Errorf("create link: database is nil")
	}
	url = strings.TrimSpace(url)
	title = strings.TrimSpace(title)
	if url == "" {
		return Link{}, fmt.Errorf("create link: url is required")
	}
	if title == "" {
		return Link{}, fmt.Errorf("create link: title is required")
	}

	now := time.Now().UTC()
	result, err := s.DB.ExecContext(ctx,
		`INSERT INTO links (url, title, created_at) VALUES (?, ?, ?)`,
		url, title, now)
	if err != nil {
		return Link{}, fmt.Errorf("create link: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Link{}, fmt.Errorf("create link id: %w", err)
	}
	return Link{ID: id, URL: url, Title: title, CreatedAt: now}, nil
}

// ListLinks returns all links ordered from newest to oldest.
func (s *Store) ListLinks(ctx context.Context) ([]Link, error) {
	if s == nil || s.DB == nil {
		return []Link{}, fmt.Errorf("list links: database is nil")
	}
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, url, title, created_at FROM links ORDER BY id DESC`)
	if err != nil {
		return []Link{}, fmt.Errorf("list links: %w", err)
	}
	defer rows.Close()

	links := []Link{}
	for rows.Next() {
		var link Link
		if err := rows.Scan(&link.ID, &link.URL, &link.Title, &link.CreatedAt); err != nil {
			return []Link{}, fmt.Errorf("list links row: %w", err)
		}
		links = append(links, link)
	}
	if err := rows.Err(); err != nil {
		return []Link{}, fmt.Errorf("list links rows: %w", err)
	}
	return links, nil
}

// DeleteLink removes a link by id.
func (s *Store) DeleteLink(ctx context.Context, id int64) error {
	if s == nil || s.DB == nil {
		return fmt.Errorf("delete link: database is nil")
	}
	if id <= 0 {
		return fmt.Errorf("delete link: invalid id %d", id)
	}
	result, err := s.DB.ExecContext(ctx, `DELETE FROM links WHERE id = ?`, id)
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
