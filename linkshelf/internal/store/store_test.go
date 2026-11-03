package store_test

import (
	"context"
	"database/sql"
	"testing"

	"linkshelf/internal/store"

	_ "modernc.org/sqlite"
)

func TestStoreCreatesAndListsItems(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := store.InitSchema(context.Background(), db); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	s := store.New(db)
	item, err := s.CreateItem(context.Background(), "first item")
	if err != nil {
		t.Fatalf("CreateItem() error = %v", err)
	}
	if item.ID == 0 || item.Text != "first item" {
		t.Fatalf("CreateItem() = %#v, want non-zero ID and text", item)
	}

	items, err := s.ListItems(context.Background())
	if err != nil {
		t.Fatalf("ListItems() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != item.ID || items[0].Text != item.Text {
		t.Fatalf("ListItems() = %#v, want [%#v]", items, item)
	}
}
