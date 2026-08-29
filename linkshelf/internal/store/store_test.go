package store_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InitSchema(db); err != nil {
		db.Close()
		t.Fatal(err)
	}
	store.DB = db
	t.Cleanup(func() {
		db.Close()
		store.DB = nil
	})
	return db
}

func TestListStartsEmptyAndOrdersNewestFirst(t *testing.T) {
	testDB(t)
	ctx := context.Background()

	links, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() on empty store returned error: %v", err)
	}
	if links == nil || len(links) != 0 {
		t.Fatalf("List() = %#v, want a non-nil empty slice", links)
	}

	first, err := store.Create(ctx, "First", "https://example.com/1")
	if err != nil {
		t.Fatalf("Create(first) returned error: %v", err)
	}
	second, err := store.Create(ctx, "Second", "http://example.com/2")
	if err != nil {
		t.Fatalf("Create(second) returned error: %v", err)
	}

	links, err = store.List(ctx)
	if err != nil {
		t.Fatalf("List() after Create returned error: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("len(List()) = %d, want 2", len(links))
	}

	// newest-first ordering and concrete field content
	if links[0].ID != second.ID {
		t.Fatalf("List()[0].ID = %d, want %d (newest first)", links[0].ID, second.ID)
	}
	if links[0].Title != "Second" {
		t.Fatalf("List()[0].Title = %q, want %q", links[0].Title, "Second")
	}
	if links[0].URL != "http://example.com/2" {
		t.Fatalf("List()[0].URL = %q, want %q", links[0].URL, "http://example.com/2")
	}
	if links[1].ID != first.ID {
		t.Fatalf("List()[1].ID = %d, want %d", links[1].ID, first.ID)
	}
	if links[1].Title != "First" {
		t.Fatalf("List()[1].Title = %q, want %q", links[1].Title, "First")
	}
	if links[1].URL != "https://example.com/1" {
		t.Fatalf("List()[1].URL = %q, want %q", links[1].URL, "https://example.com/1")
	}

	// CreatedAt must be an RFC3339 UTC timestamp
	if links[0].CreatedAt == "" {
		t.Fatal("Create() returned an empty CreatedAt")
	}
	if created, err := time.Parse(time.RFC3339, links[0].CreatedAt); err != nil || created.Location() != time.UTC {
		t.Fatalf("CreatedAt = %q, want an RFC3339 UTC timestamp", links[0].CreatedAt)
	}

	// Round-trip read-back: the link returned from Create must equal
	// the matching entry in List, including Title/URL/CreatedAt.
	if links[0].Title != second.Title || links[0].URL != second.URL || links[0].CreatedAt != second.CreatedAt {
		t.Fatalf("List()[0] = %+v, want it to match Create result %+v", links[0], second)
	}
}

func TestCreateReturnsPersistedFields(t *testing.T) {
	testDB(t)
	ctx := context.Background()

	link, err := store.Create(ctx, "Example", "https://example.com/path")
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if link.ID == 0 {
		t.Fatal("Create() returned ID == 0")
	}
	if link.Title != "Example" {
		t.Fatalf("Create().Title = %q, want %q", link.Title, "Example")
	}
	if link.URL != "https://example.com/path" {
		t.Fatalf("Create().URL = %q, want %q", link.URL, "https://example.com/path")
	}
	if link.CreatedAt == "" {
		t.Fatal("Create().CreatedAt is empty")
	}

	// Re-read by ID via List; the persisted row must round-trip with
	// the same title, url, and created_at that Create returned.
	links, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Fatalf("len(List()) = %d, want 1", len(links))
	}
	if links[0].Title != "Example" {
		t.Fatalf("persisted Title = %q, want %q", links[0].Title, "Example")
	}
	if links[0].URL != "https://example.com/path" {
		t.Fatalf("persisted URL = %q, want %q", links[0].URL, "https://example.com/path")
	}
	if links[0].CreatedAt != link.CreatedAt {
		t.Fatalf("persisted CreatedAt = %q, want %q", links[0].CreatedAt, link.CreatedAt)
	}
}

func TestCreateValidation(t *testing.T) {
	testDB(t)
	longTitle := strings.Repeat("\u754c", 201)
	cases := []struct {
		name   string
		title  string
		rawURL string
	}{
		{"empty title", "", "https://example.com"},
		{"whitespace title", " \t\n", "https://example.com"},
		{"title too long", longTitle, "https://example.com"},
		{"empty url", "Title", ""},
		{"relative url", "Title", "/path"},
		{"wrong scheme", "Title", "ftp://example.com"},
		{"scheme is case sensitive", "Title", "HTTPS://example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := store.Create(context.Background(), tc.title, tc.rawURL); err == nil {
				t.Fatal("Create() succeeded for invalid input")
			}
		})
	}
	if _, err := store.Create(context.Background(), strings.Repeat("a", 200), "https://example.com"); err != nil {
		t.Fatalf("Create() rejected a 200-rune title: %v", err)
	}
}

func TestDelete(t *testing.T) {
	testDB(t)
	ctx := context.Background()
	link, err := store.Create(ctx, "Delete me", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(ctx, link.ID); err != nil {
		t.Fatal(err)
	}
	links, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 0 {
		t.Fatalf("List() after Delete() = %#v, want empty", links)
	}
	if err := store.Delete(ctx, link.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Delete(missing) error = %v, want sql.ErrNoRows", err)
	}
}
