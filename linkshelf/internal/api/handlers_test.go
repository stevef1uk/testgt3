package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in‑memory db: %v", err)
	}
	if err := store.InitSchema(db); err != nil {
		t.Fatalf("failed to init schema: %v", err)
	}
	store.DB = db
	return db
}

func TestListLinks_Empty(t *testing.T) {
	_ = setupTestDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/links", listLinks)

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var got []store.Link
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestCreateLink_Success(t *testing.T) {
	_ = setupTestDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/links", createLink)

	body := `{"title":"Go","url":"https://golang.org"}`
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
	var created store.Link
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if created.Title != "Go" || created.URL != "https://golang.org" {
		t.Fatalf("unexpected link data: %+v", created)
	}
	if created.ID == 0 {
		t.Fatalf("expected non‑zero ID")
	}
}

func TestDeleteLink_NotFound(t *testing.T) {
	_ = setupTestDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/links/", deleteLink)

	req := httptest.NewRequest(http.MethodDelete, "/api/links/999", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rr.Code)
	}
}

func TestCreateLink_BadRequest_EmptyFields(t *testing.T) {
	_ = setupTestDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/links", createLink)

	body := `{"title":"","url":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := resp["error"]; !ok {
		t.Fatalf("expected error field in response")
	}
}

func TestCreateLink_BadRequest_InvalidJSON(t *testing.T) {
	_ = setupTestDB(t)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/links", createLink)

	body := `not json`
	req := httptest.NewRequest(http.MethodPost, "/api/links", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if _, ok := resp["error"]; !ok {
		t.Fatalf("expected error field in response")
	}
}

func TestServeStatic_Success(t *testing.T) {
	_ = setupTestDB(t)
	_ = os.Chdir(filepath.Join("..", ".."))
	t.Cleanup(func() { _ = os.Chdir(filepath.Join("internal", "api")) })

	mux := http.NewServeMux()
	mux.HandleFunc("/static/", serveStatic)

	path := filepath.Join("web", "index.html")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("skipping static test: missing %s", path)
	}

	req := httptest.NewRequest(http.MethodGet, "/static/index.html", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
