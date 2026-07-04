package store

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestCreateListDelete(t *testing.T) {
	tmpDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}
	defer tmpDB.Close()

	oldDB := DB
	DB = tmpDB
	defer func() { DB = oldDB }()

	if err := InitSchema(tmpDB); err != nil {
		t.Fatalf("InitSchema error: %v", err)
	}

	// create a link
	link, err := Create(context.Background(), "Example", "https://example.com")
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if link.Title != "Example" {
		t.Fatalf("expected title 'Example', got %q", link.Title)
	}
	if link.URL != "https://example.com" {
		t.Fatalf("expected URL 'https://example.com', got %q", link.URL)
	}
	if link.ID == 0 {
		t.Fatal("expected non-zero ID after Create")
	}

	// list should contain the link
	links, err := List(context.Background())
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].ID != link.ID {
		t.Fatalf("expected ID %d, got %d", link.ID, links[0].ID)
	}

	// delete the link
	if err := Delete(context.Background(), link.ID); err != nil {
		t.Fatalf("Delete error: %v", err)
	}

	// list should be empty
	links, err = List(context.Background())
	if err != nil {
		t.Fatalf("List error after delete: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links after delete, got %d", len(links))
	}
}

func TestCreateValidatesInput(t *testing.T) {
	tmpDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}
	defer tmpDB.Close()

	oldDB := DB
	DB = tmpDB
	defer func() { DB = oldDB }()

	if err := InitSchema(tmpDB); err != nil {
		t.Fatalf("InitSchema error: %v", err)
	}

	tests := []struct {
		name  string
		title string
		url   string
	}{
		{"empty title", "", "https://example.com"},
		{"empty url", "Example", ""},
		{"no scheme", "Example", "example.com"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Create(context.Background(), tc.title, tc.url)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestDeleteNotFoundReturnsError(t *testing.T) {
	tmpDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}
	defer tmpDB.Close()

	oldDB := DB
	DB = tmpDB
	defer func() { DB = oldDB }()

	if err := InitSchema(tmpDB); err != nil {
		t.Fatalf("InitSchema error: %v", err)
	}

	err = Delete(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent ID, got nil")
	}
	if err.Error() != "link not found" {
		t.Fatalf("expected 'link not found', got %v", err)
	}
}
