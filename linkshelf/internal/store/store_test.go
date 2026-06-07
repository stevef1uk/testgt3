package store

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func openTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}
	return db
}

// Test that Create validates input and inserts a row.
func TestCreateValid(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	DB = db

	ctx := context.Background()
	title := "Example Title"
	url := "https://example.com"
	link, err := Create(ctx, title, url)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if link.Title != title || link.URL != url {
		t.Fatalf("Created link mismatch: got %+v", link)
	}
	if link.ID == 0 {
		t.Fatalf("Expected non‑zero ID")
	}
	if _, err := time.Parse(time.RFC3339, link.CreatedAt); err != nil {
		t.Fatalf("CreatedAt not RFC3339: %v", err)
	}
}

// Test validation errors for Create.
func TestCreateValidation(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	DB = db
	ctx := context.Background()

	tests := []struct {
		name  string
		title string
		url   string
		err   string
	}{
		{"empty title", "", "https://example.com", "title must be non‑empty"},
		{"title too long", string(make([]rune, 201)), "https://example.com", "title exceeds 200"},
		{"empty url", "Title", "", "url must be non‑empty"},
		{"bad scheme", "Title", "ftp://example.com", "url must start with"},
	}
	for _, tt := range tests {
		_, err := Create(ctx, tt.title, tt.url)
		if err == nil || !strings.Contains(err.Error(), tt.err) {
			t.Fatalf("%s: expected error containing %q, got %v", tt.name, tt.err, err)
		}
	}
}

// Test List returns all inserted links in order.
func TestList(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	DB = db
	ctx := context.Background()

	links := []struct {
		title string
		url   string
	}{
		{"First", "https://first.com"},
		{"Second", "https://second.com"},
	}
	for _, l := range links {
		if _, err := Create(ctx, l.title, l.url); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}
	got, err := List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(got) != len(links) {
		t.Fatalf("List returned %d items, want %d", len(got), len(links))
	}
	for i, l := range got {
		if l.Title != links[i].title || l.URL != links[i].url {
			t.Fatalf("List item %d mismatch: got %+v, want title=%s url=%s", i, l, links[i].title, links[i].url)
		}
	}
}

// Test Delete removes a link.
func TestDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	DB = db
	ctx := context.Background()

	link, err := Create(ctx, "ToDelete", "https://todelete.com")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err := Delete(ctx, link.ID); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	// Confirm it's gone.
	list, err := List(ctx)
	if err != nil {
		t.Fatalf("List after delete failed: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("Expected empty list after delete, got %d items", len(list))
	}
}
