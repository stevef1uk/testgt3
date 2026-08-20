package api

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func testServer(t *testing.T) *http.ServeMux {
	t.Helper()
	t.Chdir(filepath.Join("..", ".."))

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := store.InitSchema(db); err != nil {
		t.Fatal(err)
	}
	store.DB = db

	mux := http.NewServeMux()
	Register(mux)
	return mux
}

func TestLinksAPIEmptyAndCreate(t *testing.T) {
	mux := testServer(t)

	res := httptest.NewRecorder()
	mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/links", nil))
	if res.Code != http.StatusOK || strings.TrimSpace(res.Body.String()) != "[]" {
		t.Fatalf("GET /api/links = %d %q", res.Code, res.Body.String())
	}

	req := httptest.NewRequest(http.MethodPost, "/api/links",
		strings.NewReader(`{"title":"Go","url":"https://go.dev"}`))
	res = httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusCreated || !strings.Contains(res.Body.String(), `"title":"Go"`) {
		t.Fatalf("POST /api/links = %d %q", res.Code, res.Body.String())
	}
}

func TestLinksAPIErrors(t *testing.T) {
	mux := testServer(t)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/links", strings.NewReader(`{"title":`))
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusBadRequest {
		t.Fatalf("bad POST status = %d", res.Code)
	}

	res = httptest.NewRecorder()
	mux.ServeHTTP(res, httptest.NewRequest(http.MethodDelete, "/api/links/999", nil))
	if res.Code != http.StatusNotFound {
		t.Fatalf("missing DELETE status = %d", res.Code)
	}
}

func TestStaticTraversalIsRejected(t *testing.T) {
	mux := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/static/%2e%2e/go.mod", nil)
	req.URL.RawPath = "/static/%2e%2e/go.mod"

	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)
	if res.Code != http.StatusNotFound {
		t.Fatalf("traversal status = %d", res.Code)
	}
}
