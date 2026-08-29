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
		t.Fatal(err)
	}
	if links == nil || len(links) != 0 {
		t.Fatalf("List() = %#v, want a non-nil empty slice", links)
	}

	first, err := store.Create(ctx, "First", "https://example.com/1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Create(ctx, "Second", "http://example.com/2")
	if err != nil {
		t.Fatal(err)
	}

	links, err = store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].ID != second.ID || links[1].ID != first.ID {
		t.Fatalf("List() = %#v, want newest-first order", links)
	}
	if links[0].CreatedAt == "" {
		t.Fatal("Create() returned an empty CreatedAt")
	}
	if created, err := time.Parse(time.RFC3339, links[0].CreatedAt); err != nil || created.Location() != time.UTC {
		t.Fatalf("CreatedAt = %q, want an RFC3339 UTC timestamp", links[0].CreatedAt)
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
