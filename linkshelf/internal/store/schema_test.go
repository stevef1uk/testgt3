package store

import (
	"database/sql"
	"reflect"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchemaIsIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := InitSchema(db); err != nil {
		t.Fatal(err)
	}
	if err := InitSchema(db); err != nil {
		t.Fatalf("calling InitSchema twice should succeed: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type = 'table' AND name = 'links'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("links table count = %d, want 1", count)
	}
}

func TestLinkFieldsAndJSONTags(t *testing.T) {
	typ := reflect.TypeOf(Link{})
	want := map[string]string
	for field, tag := range want {
		f, ok := typ.FieldByName(field)
		if !ok {
			t.Fatalf("Link is missing field %s", field)
		}
		if got := f.Tag.Get("json"); got != tag {
			t.Errorf("Link.%s json tag = %q, want %q", field, got, tag)
		}
	}
}
