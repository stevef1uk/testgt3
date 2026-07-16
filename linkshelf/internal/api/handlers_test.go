package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

// setupDB opens an in-memory SQLite database, initializes the schema,
// and wires the package-level store.DB so handler functions can use it.
func setupDB(t *testing.T) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := store.InitSchema(db); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	store.DB = db
}

func TestHandleList_Empty(t *testing.T) {
	setupDB(t)

	os.Chdir(filepath.Join("..", ".."))
	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rr := httptest.NewRecorder()
	handleList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var links []store.Link
	if err := json.NewDecoder(rr.Body).Decode(&links); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected empty list, got %d items", len(links))
	}
}

func TestHandleCreate_And_Delete(t *testing.T) {
	setupDB(t)

	os.Chdir(filepath.Join("..", ".."))
	payload := `{"title":"Test Link","url":"http://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handleCreate(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", rr.Code)
	}

	var created store.Link
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected created link to have non‑zero ID")
	}
	if created.Title != "Test Link" || created.URL != "http://example.com" {
		t.Fatalf("created link fields mismatch: %+v", created)
	}

	deletePath := "/api/links/" + strconv.FormatInt(created.ID, 10)
	delReq := httptest.NewRequest(http.MethodDelete, deletePath, nil)
	delRR := httptest.NewRecorder()
	handleDelete(delRR, delReq)

	if delRR.Code != http.StatusNoContent {
		t.Fatalf("expected status 204 No Content on delete, got %d", delRR.Code)
	}
}
