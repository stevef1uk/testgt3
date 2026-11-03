package store_test

import (
	"context"
	"database/sql"
	"testing"

	"linkshelf/internal/store"

	_ "modernc.org/sqlite"
)

func TestInitSchemaCreatesItemsTable(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := store.InitSchema(context.Background(), db); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	var name string
	if err := db.QueryRowContext(context.Background(),
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'items'",
	).Scan(&name); err != nil {
		t.Fatalf("items table was not created: %v", err)
	}
	if name != "items" {
		t.Fatalf("table name = %q, want items", name)
	}
}
