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

	"github.com/example/linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func TestHandlers(t *testing.T) {
	// Change working directory to repository root so static file serving works.
	// The test file resides in linkshelf/internal/api, so go up two levels.
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		t.Fatalf("chdir to repo root: %v", err)
	}

	// Open an in‑memory SQLite DB.
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in‑memory DB: %v", err)
	}
	// Assign the DB to the store package.
	store.DB = db
	if err := store.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	// ---- Test GET / (root) ----
	reqRoot, _ := http.NewRequest(http.MethodGet, "/", nil)
	recRoot := httptest.NewRecorder()
	handleRoot(recRoot, reqRoot)
	if recRoot.Code != http.StatusOK {
		t.Fatalf("GET /: expected 200, got %d", recRoot.Code)
	}

	// ---- Test GET /static/index.html ----
	reqStatic, _ := http.NewRequest(http.MethodGet, "/static/index.html", nil)
	recStatic := httptest.NewRecorder()
	handleStatic(recStatic, reqStatic)
	if recStatic.Code != http.StatusOK {
		t.Fatalf("GET /static/index.html: expected 200, got %d", recStatic.Code)
	}

	// ---- Test GET /api/links (empty) ----
	reqList, _ := http.NewRequest(http.MethodGet, "/api/links", nil)
	recList := httptest.NewRecorder()
	handleList(recList, reqList)
	if recList.Code != http.StatusOK {
		t.Fatalf("GET /api/links (empty): expected 200, got %d", recList.Code)
	}
	var listEmpty []store.Link
	if err := json.NewDecoder(recList.Body).Decode(&listEmpty); err != nil {
		t.Fatalf("decode empty list: %v", err)
	}
	if len(listEmpty) != 0 {
		t.Fatalf("expected empty list, got %d items", len(listEmpty))
	}

	// ---- Test POST /api/links (valid) ----
	payload := `{"title":"Example","url":"https://example.com"}`
	reqCreate, _ := http.NewRequest(http.MethodPost, "/api/links", strings.NewReader(payload))
	recCreate := httptest.NewRecorder()
	handleCreate(recCreate, reqCreate)
	if recCreate.Code != http.StatusCreated {
		t.Fatalf("POST /api/links: expected 201, got %d", recCreate.Code)
	}
	var created store.Link
	if err := json.NewDecoder(recCreate.Body).Decode(&created); err != nil {
		t.Fatalf("decode created link: %v", err)
	}
	if created.Title != "Example" || created.URL != "https://example.com" {
		t.Fatalf("created link mismatch: %+v", created)
	}

	// ---- Test GET /api/links (now contains one) ----
	reqList2, _ := http.NewRequest(http.MethodGet, "/api/links", nil)
	recList2 := httptest.NewRecorder()
	handleList(recList2, reqList2)
	if recList2.Code != http.StatusOK {
		t.Fatalf("GET /api/links (after create): expected 200, got %d", recList2.Code)
	}
	var listAfter []store.Link
	if err := json.NewDecoder(recList2.Body).Decode(&listAfter); err != nil {
		t.Fatalf("decode list after create: %v", err)
	}
	if len(listAfter) != 1 || listAfter[0].ID != created.ID {
		t.Fatalf("expected list with created link, got %+v", listAfter)
	}

	// ---- Test DELETE /api/links/{id} (valid) ----
	delPath := "/api/links/" + strconv.FormatInt(created.ID, 10)
	reqDel, _ := http.NewRequest(http.MethodDelete, delPath, nil)
	recDel := httptest.NewRecorder()
	handleDelete(recDel, reqDel)
	if recDel.Code != http.StatusNoContent {
		t.Fatalf("DELETE valid: expected 204, got %d", recDel.Code)
	}

	// ---- Test DELETE non‑existent (should 404) ----
	reqDelBad, _ := http.NewRequest(http.MethodDelete, "/api/links/9999", nil)
	recDelBad := httptest.NewRecorder()
	handleDelete(recDelBad, reqDelBad)
	if recDelBad.Code != http.StatusNotFound {
		t.Fatalf("DELETE non‑existent: expected 404, got %d", recDelBad.Code)
	}
}
