package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/testgt3/mayor/rig/linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func TestMain(m *testing.M) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()
	if err := store.InitSchema(db); err != nil {
		log.Fatalf("failed to initialize schema: %v", err)
	}
	os.Exit(m.Run())
}

func TestHandleRoot(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handleRoot(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, status)
	}
}

func TestHandleStatic(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{"valid file", "/static/index.html", http.StatusOK},
		{"path traversal", "/static/../config.yaml", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			handleStatic(w, req)

			if status := w.Code; status != tt.expected {
				t.Errorf("expected status %d, got %d for path %s", tt.expected, status, tt.path)
			}
		})
	}
}

func TestHandleListLinks(t *testing.T) {
	// Clear any existing links
	_ = store.Delete(context.Background(), 1)
	_ = store.Delete(context.Background(), 2)

	req, _ := http.NewRequest("GET", "/api/links", nil)
	w := httptest.NewRecorder()
	handleListLinks(w, req)

	if status := w.Code; status != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, status)
	}

	var links []store.Link
	if err := json.NewDecoder(w.Body).Decode(&links); err != nil {
		t.Errorf("failed to decode JSON: %v", err)
	}
	if len(links) != 0 {
		t.Errorf("expected empty list, got %d links", len(links))
	}
}

func TestHandleCreateLink(t *testing.T) {
	reqBody := `{"title":"Test","url":"https://example.com"}`
	req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handleCreateLink(w, req)

	if status := w.Code; status != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, status)
	}

	var link store.Link
	if err := json.NewDecoder(w.Body).Decode(&link); err != nil {
		t.Errorf("failed to decode JSON: %v", err)
	}
	if link.ID <= 0 {
		t.Errorf("expected positive ID, got %d", link.ID)
	}
}

func TestHandleCreateLinkInvalidJSON(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/links", strings.NewReader("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handleCreateLink(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestHandleCreateLinkMissingFields(t *testing.T) {
	reqBody := `{"title":""}`
	req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handleCreateLink(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestHandleDeleteLink(t *testing.T) {
	// Create a link first
	id, err := store.Create(context.Background(), "Test", "https://example.com")
	if err != nil {
		t.Fatalf("failed to create link: %v", err)
	}

	req, _ := http.NewRequest("DELETE", "/api/links/"+strconv.FormatInt(id.ID, 10), nil)
	w := httptest.NewRecorder()
	handleDeleteLink(w, req)

	if status := w.Code; status != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, status)
	}
}

func TestHandleDeleteLinkInvalidID(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/api/links/abc", nil)
	w := httptest.NewRecorder()
	handleDeleteLink(w, req)

	if status := w.Code; status != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, status)
	}
}

func TestHandleDeleteLinkNotFound(t *testing.T) {
	req, _ := http.NewRequest("DELETE", "/api/links/999999", nil)
	w := httptest.NewRecorder()
	handleDeleteLink(w, req)

	if status := w.Code; status != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, status)
	}
}
