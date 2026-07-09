package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

// Ensure webRoot is set before server starts (defaults to "web" relative to cwd).
var webRoot = "web"

// InitSchema calls store.InitSchema to initialize the database schema.
func InitSchema(db *sql.DB) error {
	return store.InitSchema(db)
}

// ServeIndex serves the index.html file.
func ServeIndex(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, filepath.Join(webRoot, "index.html"))
}

// ServeStatic serves static files from the web directory.
func ServeStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Reject path traversal
	reqPath := r.URL.RequestURI()
	if strings.Contains(reqPath, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	// Strip "/static/" prefix
	file := strings.TrimPrefix(r.URL.Path, "/static/")
	if file == "" {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}
	http.ServeFile(w, r, filepath.Join(webRoot, file))
}

// ListLinks returns all links as JSON.
func ListLinks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	links, err := store.List(context.Background())
	if err != nil {
		log.Printf("store.List: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Type", "application/json")
	// If empty, write exact [] without newline
	if len(links) == 0 {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(links)
	json.NewEncoder(w).Encode(links)
}

// CreateLink creates a new link from JSON body.
func CreateLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var link store.Link
	if err := json.Unmarshal(body, &link); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if link.Title == "" || link.URL == "" {
		http.Error(w, `{"error":"title and url are required"}`, http.StatusBadRequest)
		return
	}
	created, err := store.Create(context.Background(), link.Title, link.URL)
	if err != nil {
		log.Printf("store.Create: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// DeleteLink deletes a link by ID from the URL path.
func DeleteLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Extract ID from path: /api/links/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/links/")
	id, err := strconv.ParseInt(path, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, `{"error":"invalid ID"}`, http.StatusBadRequest)
		return
	}
	if err := store.Delete(context.Background(), id); err != nil {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
