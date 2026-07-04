package store

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite" // pure-Go SQLite driver
)

func TestInitSchema(t *testing.T) {
	tmpDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}
	defer tmpDB.Close()

	oldDB := DB
	DB = tmpDB
	defer func() { DB = oldDB }()

	if err := InitSchema(tmpDB); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	// ensure the table exists and is empty
	lst, err := List(context.Background())
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(lst) != 0 {
		t.Fatalf("expected empty list after InitSchema, got %v", lst)
	}
}
