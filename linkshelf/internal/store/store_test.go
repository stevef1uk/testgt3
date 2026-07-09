package store

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupStore(t *testing.T) (*sql.DB, context.Context) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}
	// Save original DB, set to test DB, restore after test.
	origDB := DB
	DB = db
	t.Cleanup(func() { DB = origDB; db.Close() })
	return db, context.Background()
}

func TestList_empty(t *testing.T) {
	_, ctx := setupStore(t)
	links, err := List(ctx)
	if err != nil {
		t.Fatalf("List on empty store: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links, got %d", len(links))
	}
}

func TestCreateAndList(t *testing.T) {
	_, ctx := setupStore(t)
	// Create two links
	l1, err := Create(ctx, "Example", "https://example.com")
	if err != nil {
		t.Fatalf("Create first link: %v", err)
	}
	if l1.Title != "Example" || l1.URL != "https://example.com" || l1.CreatedAt == "" || l1.ID == 0 {
		t.Fatalf("unexpected first link: %+v", l1)
	}
	l2, err := Create(ctx, "Go", "https://go.dev")
	if err != nil {
		t.Fatalf("Create second link: %v", err)
	}
	links, err := List(ctx)
	if err != nil {
		t.Fatalf("List after creates: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d: %+v", len(links), links)
	}
	if links[0].ID != l1.ID || links[1].ID != l2.ID {
		t.Fatalf("list order or IDs mismatch: %+v", links)
	}
}

func TestCreate_validation_title_empty(t *testing.T) {
	_, ctx := setupStore(t)
	_, err := Create(ctx, "", "https://example.com")
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}

func TestCreate_validation_title_too_long(t *testing.T) {
	_, ctx := setupStore(t)
	longTitle := string(make([]rune, 201))
	_, err := Create(ctx, longTitle, "https://example.com")
	if err == nil {
		t.Fatal("expected error for title >200 runes")
	}
}

func TestCreate_validation_url(t *testing.T) {
	_, ctx := setupStore(t)
	tests := []struct {
		url string
		ok  bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"ftp://example.com", false},
		{"example.com", false},
		{"", false},
	}
	for _, tc := range tests {
		_, err := Create(ctx, "Title", tc.url)
		if tc.ok && err != nil {
			t.Errorf("expected OK for url %q, got error: %v", tc.url, err)
		}
		if !tc.ok && err == nil {
			t.Errorf("expected error for url %q, got OK", tc.url)
		}
	}
}

func TestDelete(t *testing.T) {
	_, ctx := setupStore(t)
	l, err := Create(ctx, "Del Me", "https://example.com")
	if err != nil {
		t.Fatalf("Create before delete: %v", err)
	}
	if err := Delete(ctx, l.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	links, err := List(ctx)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links after delete, got %d", len(links))
	}
}

func TestDelete_nonexistent(t *testing.T) {
	_, ctx := setupStore(t)
	if err := Delete(ctx, 9999); err != nil {
		t.Fatalf("Delete nonexistent ID: %v", err)
	}
}

func TestList_nilDB(t *testing.T) {
	orig := DB
	DB = nil
	t.Cleanup(func() { DB = orig })
	_, err := List(context.Background())
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}
}

func TestCreate_nilDB(t *testing.T) {
	orig := DB
	DB = nil
	t.Cleanup(func() { DB = orig })
	_, err := Create(context.Background(), "x", "https://x.com")
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}
}

func TestDelete_nilDB(t *testing.T) {
	orig := DB
	DB = nil
	t.Cleanup(func() { DB = orig })
	err := Delete(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when DB is nil")
	}
}
