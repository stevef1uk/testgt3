package store

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := InitSchema(db); err != nil {
		t.Fatal(err)
	}
	DB = db
	return db
}

func TestCreateAndList(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	link, err := Create(ctx, "Example", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	if link.Title != "Example" {
		t.Errorf("expected title Example, got %s", link.Title)
	}
	if link.URL != "https://example.com" {
		t.Errorf("expected URL https://example.com, got %s", link.URL)
	}
	if link.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if link.CreatedAt == "" {
		t.Error("expected non-empty CreatedAt")
	}

	links, err := List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].ID != link.ID {
		t.Errorf("expected ID %d, got %d", link.ID, links[0].ID)
	}
}

func TestCreateValidation(t *testing.T) {
	setupTestDB(t)

	_, err := Create(context.Background(), "", "https://valid.com")
	if err != nil {
		t.Error("expected no error for empty title (no length constraint on empty)")
	}

	// title longer than 200 runes
	longTitle := string(make([]rune, 201))
	_, err = Create(context.Background(), longTitle, "https://example.com")
	if err == nil {
		t.Error("expected error for title >200 runes")
	}

	// url not starting with http/https
	_, err = Create(context.Background(), "Test", "ftp://bad.com")
	if err == nil {
		t.Error("expected error for non-http/https URL")
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	link, err := Create(ctx, "ToDelete", "https://delete.me")
	if err != nil {
		t.Fatal(err)
	}

	if err := Delete(ctx, link.ID); err != nil {
		t.Fatal(err)
	}

	// Verify deletion
	links, err := List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 0 {
		t.Errorf("expected 0 links after delete, got %d", len(links))
	}

	// Delete non-existent should error
	if err := Delete(ctx, 99999); err == nil {
		t.Error("expected error deleting non-existent link")
	}
}
