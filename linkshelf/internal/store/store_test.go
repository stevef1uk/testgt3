package store

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func testStore(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := InitSchema(db); err != nil {
		t.Fatal(err)
	}
	DB = db
	t.Cleanup(func() { DB = nil })
	return db
}

func TestStoreCRUD(t *testing.T) {
	testStore(t)
	ctx := context.Background()

	empty, err := List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if empty == nil || len(empty) != 0 {
		t.Fatalf("expected non-nil empty list, got %#v", empty)
	}

	first, err := Create(ctx, "First", "https://first.example")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	second, err := Create(ctx, "Second", "http://second.example")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == 0 || second.ID <= first.ID {
		t.Fatalf("unexpected IDs: first=%d second=%d", first.ID, second.ID)
	}
	if first.CreatedAt.Location() != time.UTC || first.CreatedAt.Format(time.RFC3339) == "" {
		t.Fatalf("unexpected timestamp: %v", first.CreatedAt)
	}

	links, err := List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].ID != second.ID || links[1].ID != first.ID {
		t.Fatalf("expected descending IDs, got %#v", links)
	}

	if err := Delete(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if err := Delete(ctx, first.ID); err == nil {
		t.Fatal("expected missing delete error")
	}
}

func TestCreateValidation(t *testing.T) {
	testStore(t)
	ctx := context.Background()
	tests := []struct {
		name  string
		title string
		url   string
	}{
		{"empty title", "", "https://example.com"},
		{"long title", strings.Repeat("a", 201), "https://example.com"},
		{"empty url", "Example", ""},
		{"invalid scheme", "Example", "ftp://example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Create(ctx, tt.title, tt.url); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
