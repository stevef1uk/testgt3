package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/testgt3/mayor/rig/linkshelf/internal/store"
)

// handleRoot serves the index.html file from the web directory.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	webRoot := filepath.Join(".", "web")
	indexPath := filepath.Join(webRoot, "index.html")
	http.ServeFile(w, r, indexPath)
}

// handleStatic serves static files from the web directory.
// It rejects any request containing ".." to prevent path traversal.
func handleStatic(w http.ResponseWriter, r *http.Request) {
	// Expected pattern: /static/{file}
	uri := r.URL.Path
	// Trim the "/static/" prefix
	if !strings.HasPrefix(uri, "/static/") {
		http.NotFound(w, r)
		return
	}
	relPath := strings.TrimPrefix(uri, "/static/")
	// Reject path traversal
	if strings.Contains(relPath, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	webRoot := filepath.Join(".", "web")
	fullPath := filepath.Join(webRoot, relPath)

	// Ensure the file exists and is not a directory
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, fullPath)
}

// handleListLinks returns the list of all stored links as JSON.
func handleListLinks(w http.ResponseWriter, r *http.Request) {
	links, err := store.List(r.Context())
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(links)
}

// handleCreateLink creates a new link from JSON payload.
// On validation error it returns 400 with a JSON error object.
func handleCreateLink(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
		return
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	if payload.Title == "" || payload.URL == "" {
		http.Error(w, `{"error":"title and url are required"}`, http.StatusBadRequest)
		return
	}
	link, err := store.Create(r.Context(), payload.Title, payload.URL)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

// handleDeleteLink deletes a link by its ID extracted from the URL path.
// URL pattern: /api/links/{id}
func handleDeleteLink(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/api/links/") {
		http.NotFound(w, r)
		return
	}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/links/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := store.Delete(r.Context(), id); err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
