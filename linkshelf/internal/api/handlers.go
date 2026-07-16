package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	// Serve the main HTML page
	http.ServeFile(w, r, filepath.Join("web", "index.html"))
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	// Reject any path traversal attempts
	if strings.Contains(r.URL.Path, "..") {
		http.NotFound(w, r)
		return
	}
	// Expected format: /static/{file}
	relPath := strings.TrimPrefix(r.URL.Path, "/static/")
	if relPath == "" {
		http.NotFound(w, r)
		return
	}
	// Prevent escaped sequences from being interpreted as paths
	if u, err := url.PathUnescape(relPath); err == nil && strings.Contains(u, "..") {
		http.NotFound(w, r)
		return
	}
	absPath := filepath.Join("web", relPath)
	if _, err := os.Stat(absPath); err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, absPath)
}

func handleList(w http.ResponseWriter, r *http.Request) {
	links, err := store.List()
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	link := store.Link{
		Title: payload.Title,
		URL:   payload.URL,
	}
	id, err := store.Create(link)
	if err != nil {
		resp := map[string]string{"error": err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}
	// Populate ID field for response.
	link.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	// Expected format: /api/links/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/links/")
	id64, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
		return
	}
	if err := store.Delete(id64); err != nil {
		resp := map[string]string{"error": err.Error()}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(resp)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
