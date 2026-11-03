package store_test

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"linkshelf/internal/store"

	_ "modernc.org/sqlite"
)

func testDB(t *testing.T) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InitSchema(db); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	store.DB = db
	t.Cleanup(func() {
		store.DB = nil
		_ = db.Close()
	})
}

func TestStoreListCreateDelete(t *testing.T) {
	testDB(t)
	ctx := context.Background()

	links, err := store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if links == nil {
		t.Fatal("List returned nil slice")
	}

	first, err := store.Create(ctx, "First", "https://example.com/1")
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.Create(ctx, "Second", "https://example.com/2")
	if err != nil {
		t.Fatal(err)
	}
	if second.ID <= first.ID || second.CreatedAt == "" {
		t.Fatalf("Create results = %#v, %#v", first, second)
	}
	if _, err := time.Parse(time.RFC3339, second.CreatedAt); err != nil {
		t.Fatalf("CreatedAt = %q: %v", second.CreatedAt, err)
	}

	links, err = store.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].ID != second.ID || links[1].ID != first.ID {
		t.Fatalf("List = %#v, want newest first", links)
	}

	if err := store.Delete(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(ctx, first.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("missing Delete error = %v, want ErrNotFound", err)
	}
}

func TestStoreCreateValidation(t *testing.T) {
	testDB(t)
	cases := []struct {
		name, title, url string
	}{
		{"empty title", "", "https://example.com"},
		{"whitespace title", " \t\n", "https://example.com"},
		{"long title", strings.Repeat("x", 201), "https://example.com"},
		{"empty url", "Title", ""},
		{"relative url", "Title", "/path"},
		{"wrong scheme", "Title", "ftp://example.com"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := store.Create(context.Background(), tc.title, tc.url); err == nil {
				t.Fatal("Create accepted invalid input")
			}
		})
	}
}
