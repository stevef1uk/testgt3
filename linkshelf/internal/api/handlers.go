package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

// Link represents a bookmark.
type Link = store.Link

// HandleGetRoot serves the index.html file.
func HandleGetRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join("web", "index.html"))
}

// HandleGetStatic serves files from the web/ directory.
func HandleGetStatic(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.Contains(path, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	http.StripPrefix("/static/", http.FileServer(http.Dir("web"))).ServeHTTP(w, r)
}

// HandleListLinks returns all links as JSON.
func HandleListLinks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	links, err := store.List(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

// HandleCreateLink creates a new link from POST JSON body.
func HandleCreateLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var input struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	link, err := store.Create(ctx, input.Title, input.URL)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(link)
}

// HandleDeleteLink deletes a link by ID from POST path.
func HandleDeleteLink(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Expect pattern: DELETE /api/links/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		http.NotFound(w, r)
		return
	}
	idStr := parts[len(parts)-1]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, `{"error":"invalid id"}`, http.StatusNotFound)
		return
	}
	if err := store.Delete(ctx, id); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
