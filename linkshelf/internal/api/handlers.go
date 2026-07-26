package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

// handleRoot serves the main HTML page.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	// The web assets are placed under ./web relative to the module root.
	// When tests run from the package directory we first changed the working
	// directory to the module root (see handlers_test.go init), so this path
	// resolves correctly.
	http.ServeFile(w, r, filepath.Join("web", "index.html"))
}

// handleStatic serves static files from the ./web directory.
func handleStatic(w http.ResponseWriter, r *http.Request) {
	// Trim the leading '/' and ensure the path stays within the web directory.
	trimmed := strings.TrimPrefix(r.URL.Path, "/")
	if trimmed == "" {
		http.NotFound(w, r)
		return
	}
	// Only allow files inside the web folder.
	http.ServeFile(w, r, filepath.Join("web", trimmed))
}

// handleList returns all stored links as JSON.
// Expected method: GET.
func handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	links, err := store.List()
	if err != nil {
		http.Error(w, `{"error":"failed to list links"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(links); err != nil {
		http.Error(w, `{"error":"failed to encode links"}`, http.StatusInternalServerError)
	}
}

// handleCreate adds a new link.
// Expected method: POST with JSON body { "title": "...", "url": "..." }.
func handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if payload.Title == "" || payload.URL == "" {
		http.Error(w, `{"error":"title and url required"}`, http.StatusBadRequest)
		return
	}
	id, err := store.Create(payload.Title, payload.URL)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}
	// Build the response Link using the returned ID and the original payload.
	link := store.Link{
		ID:    id,
		Title: payload.Title,
		URL:   payload.URL,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(link)
}

// handleDelete removes a link by ID.
// Expected method: DELETE on path /api/links/{id}.
func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Expected URL: /api/links/{id}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(parts[2])
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := store.Delete(int64(id)); err != nil {
		// Delete may return an error if the row does not exist.
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
