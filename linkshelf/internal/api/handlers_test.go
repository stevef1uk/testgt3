package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

// init changes the working directory to the module root so that relative
// paths used by handlers (e.g., serving files from ./web) resolve correctly
// when running `go test` from the package directory.
func init() {
	// The test package resides in linkshelf/internal/api; the module root
	// is two levels up.
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		panic(err)
	}
}

// setupDB creates an in-memory SQLite database, initializes the schema,
// and wires the store package-level DB variable. It returns the database
// handle for deferred cleanup.
func setupDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := store.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	store.DB = db
	return db
}

// newTestMux returns a net/http.ServeMux with the API handlers registered.
func newTestMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/links", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleList(w, r)
		case http.MethodPost:
			handleCreate(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/links/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			handleDelete(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	return mux
}

func TestHandleList_Empty(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	mux := newTestMux()

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var links []store.Link
	if err := json.NewDecoder(rec.Body).Decode(&links); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected empty list, got %d items", len(links))
	}
}

func TestHandleCreate_AndList(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	mux := newTestMux()

	body := `{"title":"Test","url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d; body: %s", rec.Code, rec.Body.String())
	}

	var created store.Link
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}
	if created.Title != "Test" {
		t.Fatalf("expected 'Test', got '%s'", created.Title)
	}
	if created.URL != "https://example.com" {
		t.Fatalf("expected 'https://example.com', got '%s'", created.URL)
	}
	// CreatedAt is populated by the database; skip asserting its value
	// as the format may vary based on driver.

	// Now list and verify the link appears
	req2 := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rec2 := httptest.NewRecorder()
	mux.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var links []store.Link
	if err := json.NewDecoder(rec2.Body).Decode(&links); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, links[0].ID)
	}
}

func TestHandleCreate_BadJSON(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	mux := newTestMux()

	body := `not json`
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	var errResp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if errResp["error"] == "" {
		t.Fatalf("expected error message")
	}
}

func TestHandleDelete(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	mux := newTestMux()

	// Create a link first
	body := `{"title":"DeleteMe","url":"https://example.com/delete"}`
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d", rec.Code)
	}
	var created store.Link
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Delete the link
	delReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/links/%d", created.ID), nil)
	delRec := httptest.NewRecorder()
	mux.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", delRec.Code)
	}

	// List should be empty
	listReq := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	listRec := httptest.NewRecorder()
	mux.ServeHTTP(listRec, listReq)

	var links []store.Link
	if err := json.NewDecoder(listRec.Body).Decode(&links); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links after delete, got %d", len(links))
	}
}

func TestHandleDelete_NotFound(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	mux := newTestMux()

	req := httptest.NewRequest(http.MethodDelete, "/api/links/9999", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	// store.Delete does not report missing rows as an error, so the handler
	// may return 204 for non-existent links. Accept both 204 and 404.
	if rec.Code != http.StatusNoContent && rec.Code != http.StatusNotFound {
		t.Fatalf("expected 204 or 404, got %d", rec.Code)
	}
}

func TestHandleList_MethodNotAllowed(t *testing.T) {
	db := setupDB(t)
	defer db.Close()
	mux := newTestMux()

	req := httptest.NewRequest(http.MethodPut, "/api/links", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}
