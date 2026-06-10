package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"linkshelf/internal/store"

	"github.com/go-chi/chi/v5"
)

// initStore creates an in‑memory SQLite DB, runs the schema migration, and
// assigns the DB to the store package variable used by the handlers.
func initStore(t *testing.T) {
	// Change to the repository root so relative paths (e.g. for web assets) resolve correctly.
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		t.Fatalf("chdir to repo root: %v", err)
	}
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in‑memory DB: %v", err)
	}
	if err := store.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	// Assign the DB to the store package variable that handlers use.
	store.DB = db
}

// setWebRoot configures the package‑level webRoot variable to the repository's web directory.
func setWebRoot(t *testing.T) {
	dir, err := filepath.Abs(filepath.Join("..", "..", "web"))
	if err != nil {
		t.Fatalf("resolve web dir: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("web dir missing: %v", err)
	}
	webRoot = dir
}

func TestServeIndex(t *testing.T) {
	setWebRoot(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	serveIndex(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Fatalf("expected non‑empty index.html response")
	}
}

func TestServeStatic(t *testing.T) {
	setWebRoot(t)
	// Successful static file request.
	req := httptest.NewRequest(http.MethodGet, "/static/app.js", nil)
	rr := httptest.NewRecorder()
	serveStatic(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for static file, got %d", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Fatalf("expected non‑empty static file response")
	}
	// Path traversal should be rejected.
	req = httptest.NewRequest(http.MethodGet, "/static/../go.mod", nil)
	rr = httptest.NewRecorder()
	serveStatic(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for path traversal, got %d", rr.Code)
	}
}

func TestHandleGetLinks(t *testing.T) {
	initStore(t)
	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rr := httptest.NewRecorder()
	handleGetLinks(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", rr.Code)
	}
	var got []store.Link
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(got))
	}
}

func TestHandlePostLink(t *testing.T) {
	initStore(t)
	payload := map[string]string{
		"title": "Example",
		"url":   "https://example.com",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handlePostLink(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", rr.Code)
	}
	var created store.Link
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("decode created link: %v", err)
	}
	if created.Title != payload["title"] || created.URL != payload["url"] {
		t.Fatalf("link fields mismatch: got %+v, want %+v", created, payload)
	}
	if created.ID == 0 {
		t.Fatalf("expected non‑zero ID for created link")
	}
}

func TestHandleDeleteLink(t *testing.T) {
	initStore(t)
	// Create a link to delete.
	ln, err := store.Create(context.Background(), "to‑delete", "https://delete.me")
	if err != nil {
		t.Fatalf("store.Create failed: %v", err)
	}
	// Build request with chi URL param.
	req := httptest.NewRequest(http.MethodDelete, "/api/links/"+strconv.FormatInt(ln.ID, 10), nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(ln.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()
	handleDeleteLink(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", rr.Code)
	}
	// Verify the link was removed by checking List returns empty.
	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/links", nil)
	handleGetLinks(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for List after delete, got %d", rr.Code)
	}
	var after []store.Link
	if err := json.NewDecoder(rr.Body).Decode(&after); err != nil {
		t.Fatalf("decode List after delete: %v", err)
	}
	if len(after) != 0 {
		t.Fatalf("expected empty slice after delete, got %d items", len(after))
	}
}

func TestRegisterHandlers(t *testing.T) {
	// Ensure RegisterHandlers does not panic.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("RegisterHandlers panicked: %v", r)
		}
	}()
	RegisterHandlers()
	// Simple sanity: root should return 200.
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	http.DefaultServeMux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for root after RegisterHandlers, got %d", rr.Code)
	}
}
