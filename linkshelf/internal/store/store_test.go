package store

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestStoreCreatesListsAndDeletesLinks(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := InitSchema(db); err != nil {
		t.Fatal(err)
	}
	s := New(db)

	created, err := s.CreateLink(context.Background(), "https://example.com", "Example")
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.URL != "https://example.com" || created.Title != "Example" {
		t.Fatalf("unexpected created link: %#v", created)
	}

	links, err := s.ListLinks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].ID != created.ID {
		t.Fatalf("unexpected links: %#v", links)
	}

	if err := s.DeleteLink(context.Background(), created.ID); err != nil {
		t.Fatal(err)
	}
	links, err = s.ListLinks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if links == nil || len(links) != 0 {
		t.Fatalf("expected an empty, non-nil list: %#v", links)
	}
}
