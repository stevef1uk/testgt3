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

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	if err := store.InitSchema(db); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	origDB := store.DB
	store.DB = db
	cleanup := func() {
		store.DB = origDB
		db.Close()
	}
	return db, cleanup
}

func TestListLinks_empty(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	w := httptest.NewRecorder()
	ListLinks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var links []store.Link
	if err := json.NewDecoder(w.Body).Decode(&links); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links, got %d", len(links))
	}
}

func TestListLinks_method_not_allowed(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/links", nil)
	w := httptest.NewRecorder()
	ListLinks(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestListLinks_with_items(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	store.Create(context.Background(), "First", "https://first.com")
	store.Create(context.Background(), "Second", "https://second.com")

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	w := httptest.NewRecorder()
	ListLinks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var links []store.Link
	if err := json.NewDecoder(w.Body).Decode(&links); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(links))
	}
	if links[0].Title != "First" || links[1].Title != "Second" {
		t.Fatalf("unexpected order or titles: %+v", links)
	}
}

func TestCreateLink_success(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	body := `{"title":"Example","url":"https://example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	CreateLink(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var link store.Link
	if err := json.NewDecoder(w.Body).Decode(&link); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if link.Title != "Example" || link.URL != "https://example.com" || link.ID == 0 {
		t.Fatalf("unexpected link: %+v", link)
	}
}

func TestCreateLink_invalid_json(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(`{bad`))
	w := httptest.NewRecorder()
	CreateLink(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateLink_missing_fields(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	body := `{"title":"","url":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	CreateLink(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateLink_method_not_allowed(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	w := httptest.NewRecorder()
	CreateLink(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestDeleteLink_success(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	link, err := store.Create(context.Background(), "Delete Me", "https://example.com")
	if err != nil {
		t.Fatalf("store.Create: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/links/"+strconv.FormatInt(link.ID, 10), nil)
	w := httptest.NewRecorder()
	DeleteLink(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}

	links, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("store.List: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links after delete, got %d", len(links))
	}
}

func TestDeleteLink_invalid_id(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/links/abc", nil)
	w := httptest.NewRecorder()
	DeleteLink(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteLink_nonexistent(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodDelete, "/api/links/9999", nil)
	w := httptest.NewRecorder()
	DeleteLink(w, req)

	// store.Delete returns nil for non-existent IDs (not an error)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteLink_method_not_allowed(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	_, cleanup := setupTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/api/links/1", nil)
	w := httptest.NewRecorder()
	DeleteLink(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestServeIndex(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	origWebRoot := WebRoot
	WebRoot, _ = os.MkdirTemp("", "webroot")
	t.Cleanup(func() { os.RemoveAll(WebRoot); WebRoot = origWebRoot })

	indexContent := "<html><body>Test</body></html>"
	if err := os.WriteFile(filepath.Join(WebRoot, "index.html"), []byte(indexContent), 0644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	ServeIndex(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != indexContent {
		t.Fatalf("expected body %q, got %q", indexContent, w.Body.String())
	}
}

func TestServeIndex_method_not_allowed(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()
	ServeIndex(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestServeStatic(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	origWebRoot := WebRoot
	WebRoot, _ = os.MkdirTemp("", "webroot")
	t.Cleanup(func() { os.RemoveAll(WebRoot); WebRoot = origWebRoot })

	cssContent := "body { color: red; }"
	if err := os.WriteFile(filepath.Join(WebRoot, "style.css"), []byte(cssContent), 0644); err != nil {
		t.Fatalf("write style.css: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/static/style.css", nil)
	w := httptest.NewRecorder()
	ServeStatic(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != cssContent {
		t.Fatalf("expected body %q, got %q", cssContent, w.Body.String())
	}
}

func TestServeStatic_path_traversal(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	req := httptest.NewRequest(http.MethodGet, "/static/../etc/passwd", nil)
	w := httptest.NewRecorder()
	ServeStatic(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestServeStatic_method_not_allowed(t *testing.T) {
	os.Chdir(filepath.Join("..", ".."))
	origCwd, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(origCwd) })

	req := httptest.NewRequest(http.MethodPost, "/static/style.css", nil)
	w := httptest.NewRecorder()
	ServeStatic(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}
