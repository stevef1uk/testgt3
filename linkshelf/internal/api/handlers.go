package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"

	"github.com/go-chi/chi/v5"
)

// webRoot points to the directory containing static web assets.
var webRoot = "web"

// writeJSONError writes an error message as a JSON object with the given status code.
func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, msg)
}

// serveIndex serves the main index.html file.
func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(webRoot, "index.html"))
}

// serveStatic serves static files from the web directory.
func serveStatic(w http.ResponseWriter, r *http.Request) {
	uri := r.URL.Path
	// Prevent path traversal attacks.
	if strings.Contains(uri, "..") {
		http.NotFound(w, r)
		return
	}
	relPath := strings.TrimPrefix(uri, "/static/")
	if relPath == "" {
		http.NotFound(w, r)
		return
	}
	absPath := filepath.Join(webRoot, relPath)
	http.ServeFile(w, r, absPath)
}

// handleGetLinks retrieves all links from the store.
func handleGetLinks(w http.ResponseWriter, r *http.Request) {
	links, err := store.List(r.Context())
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

// handlePostLink creates a new link in the store.
func handlePostLink(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}
	created, err := store.Create(r.Context(), payload.Title, payload.URL)
	if err != nil {
		// Validation errors are client errors; other errors are server errors.
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must") {
			writeJSONError(w, err.Error(), http.StatusBadRequest)
		} else {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// handleDeleteLink removes a link by ID.
func handleDeleteLink(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSONError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := store.Delete(r.Context(), int64(id)); err != nil {
		// Not‑found or other DB error.
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RegisterHandlers registers all HTTP handlers on the default ServeMux.
func RegisterHandlers() {
	// Root and static assets.
	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/static/", serveStatic)

	// API endpoints.
	http.HandleFunc("/api/links", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetLinks(w, r)
		case http.MethodPost:
			handlePostLink(w, r)
		default:
			writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// DELETE endpoint expects an ID as the final path component.
	http.HandleFunc("/api/links/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		handleDeleteLink(w, r)
	})
}
