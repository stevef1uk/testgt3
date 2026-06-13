package api

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/example/linkshelf/internal/store"
)

const webRoot = "web"

// handleRoot serves the UI index page.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(webRoot, "index.html"))
}

// handleStatic serves static files under the web directory.
// It rejects any path containing ".." to prevent directory traversal.
func handleStatic(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	if !strings.HasPrefix(uri, "/static/") {
		http.NotFound(w, r)
		return
	}
	rel := strings.TrimPrefix(uri, "/static/")
	if strings.Contains(rel, "..") {
		http.NotFound(w, r)
		return
	}
	absPath := filepath.Join(webRoot, rel)

	// Serve the file manually to avoid unexpected redirects.
	f, err := os.Open(absPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	// Detect content type.
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	w.Header().Set("Content-Type", contentType)
	// Reset file offset after reading header bytes.
	if _, err := f.Seek(0, 0); err != nil {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}

// handleList returns the list of links as JSON.
func handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	links, err := store.List(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(links)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

// handleCreate creates a new link from JSON payload.
func handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	var payload struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if payload.Title == "" {
		payload.Title = payload.URL
	}
	link, err := store.Create(r.Context(), payload.Title, payload.URL)
	if err != nil {
		resp := map[string]string{"error": err.Error()}
		b, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)
		return
	}
	b, _ := json.Marshal(link)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(b)
}

// handleDelete deletes a link by ID from the URL path.
func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Expected path: /api/links/{id}
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/links/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := store.Delete(r.Context(), id); err != nil {
		resp := map[string]string{"error": err.Error()}
		b, _ := json.Marshal(resp)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(b)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
