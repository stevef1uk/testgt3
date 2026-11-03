package store_test

import (
	"context"
	"database/sql"
	"testing"

	"linkshelf/internal/store"

	_ "modernc.org/sqlite"
)

func TestInitSchemaCreatesLinksTable(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := store.InitSchema(db); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	var name string
	if err := db.QueryRowContext(context.Background(),
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'links'",
	).Scan(&name); err != nil {
		t.Fatalf("items table was not created: %v", err)
	}
	if name != "links" {
		t.Fatalf("table name = %q, want links", name)
	}
}
