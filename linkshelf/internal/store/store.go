package store

import (
	"context"
	"database/sql"
	"fmt"
)

type Item struct {
	ID   int64  `json:"id"`
	Text string `json:"text"`
}

type Store struct {
	db *sql.DB
}

func New(db *sql.DB) *Store {
	return &Store{db: db}
}

func InitSchema(ctx context.Context, db *sql.DB) error {
	const query = `
		CREATE TABLE IF NOT EXISTS items (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			text TEXT NOT NULL
		)`
	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("initialize schema: %w", err)
	}
	return nil
}

func (s *Store) CreateItem(ctx context.Context, text string) (Item, error) {
	result, err := s.db.ExecContext(ctx, "INSERT INTO items (text) VALUES (?)", text)
	if err != nil {
		return Item{}, fmt.Errorf("create item: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return Item{}, fmt.Errorf("get created item ID: %w", err)
	}
	return Item{ID: id, Text: text}, nil
}

func (s *Store) ListItems(ctx context.Context) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, text FROM items ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var item Item
		if err := rows.Scan(&item.ID, &item.Text); err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}
	return items, nil
}
