package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var DB *sql.DB

func List(ctx context.Context) ([]Link, error) {
	if DB == nil {
		return nil, errors.New("store database is not initialized")
	}

	rows, err := DB.QueryContext(ctx, `
		SELECT id, title, url, created_at
		FROM links
		ORDER BY id DESC`)
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

func Create(ctx context.Context, title, url string) (Link, error) {
	if DB == nil {
		return Link{}, errors.New("store database is not initialized")
	}
	if strings.TrimSpace(title) == "" {
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

	result, err := DB.ExecContext(ctx,
		`INSERT INTO links (title, url, created_at) VALUES (?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))`,
		title, url)
	if err != nil {
		return Link{}, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Link{}, err
	}

	var link Link
	err = DB.QueryRowContext(ctx,
		`SELECT id, title, url, created_at FROM links WHERE id = ?`, id).
		Scan(&link.ID, &link.Title, &link.URL, &link.CreatedAt)
	if err != nil {
		return Link{}, fmt.Errorf("read created link: %w", err)
	}
	return link, nil
}

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
