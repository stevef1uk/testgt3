package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func newTestDB(t *testing.T) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	DB = db
}

func TestCreate_AndList(t *testing.T) {
	newTestDB(t)
	ctx := context.Background()

	l, err := Create(ctx, "Example", "https://example.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if l.ID == 0 || l.Title != "Example" || l.URL != "https://example.com" || l.CreatedAt == "" {
		t.Fatalf("unexpected link: %+v", l)
	}

	list, err := List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 link, got %d", len(list))
	}
	if list[0].ID != l.ID {
		t.Fatalf("expected id %d, got %d", l.ID, list[0].ID)
	}
}

func TestList_EmptyReturnsNonNilSlice(t *testing.T) {
	newTestDB(t)
	list, err := List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if list == nil {
		t.Fatalf("expected non-nil empty slice")
	}
	if len(list) != 0 {
		t.Fatalf("expected empty, got %d", len(list))
	}
}

func TestCreate_ValidationErrors(t *testing.T) {
	newTestDB(t)
	ctx := context.Background()

	cases := []struct {
		title, url string
	}{
		{"", "https://example.com"},
		{"   ", "https://example.com"},
		{"Example", ""},
		{"Example", "ftp://example.com"},
		{"Example", "example.com"},
	}
	for _, c := range cases {
		if _, err := Create(ctx, c.title, c.url); err == nil {
			t.Errorf("expected error for title=%q url=%q", c.title, c.url)
		}
	}
}

func TestCreate_TitleTooLong(t *testing.T) {
	newTestDB(t)
	long := make([]byte, 201)
	for i := range long {
		long[i] = 'a'
	}
	if _, err := Create(context.Background(), string(long), "https://example.com"); err == nil {
		t.Fatalf("expected error for 201-rune title")
	}
}

func TestDelete_ExistingAndMissing(t *testing.T) {
	newTestDB(t)
	ctx := context.Background()

	l, err := Create(ctx, "ToDelete", "https://example.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := Delete(ctx, l.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := Delete(ctx, l.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound on second delete, got %v", err)
	}
}
