package store

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCRUDValidatesAndOrdersLinks(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := InitSchema(db); err != nil {
		t.Fatal(err)
	}
	DB = db

	ctx := context.Background()
	if _, err := Create(ctx, "", "https://example.com"); err == nil {
		t.Fatal("Create accepted an empty title")
	}
	if _, err := Create(ctx, "Example", ""); err == nil {
		t.Fatal("Create accepted an empty URL")
	}

	first, err := Create(ctx, "First", "https://first.example")
	if err != nil {
		t.Fatal(err)
	}
	second, err := Create(ctx, "Second", "https://second.example")
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == 0 || second.ID == 0 || first.CreatedAt == "" || second.CreatedAt == "" {
		t.Fatalf("Create returned incomplete links: %#v %#v", first, second)
	}

	links, err := List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 2 || links[0].ID != second.ID || links[1].ID != first.ID {
		t.Fatalf("List order = %#v, want newest first", links)
	}

	if err := Delete(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	links, err = List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].ID != second.ID {
		t.Fatalf("after Delete = %#v", links)
	}
	if err := Delete(ctx, first.ID); err == nil {
		t.Fatal("Delete unexpectedly succeeded for a missing link")
	}
}
