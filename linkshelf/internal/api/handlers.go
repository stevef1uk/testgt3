package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/static/", staticFileHandler)
	mux.HandleFunc("/api/links", linksHandler)
	mux.HandleFunc("/api/links/", deleteLinkHandler)
}

func linksHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listLinksHandler(w, r)
	case http.MethodPost:
		createLinkHandler(w, r)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func listLinksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	links, err := store.List(context.Background())
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if links == nil {
		links = make([]store.Link, 0)
	}
	writeJSON(w, http.StatusOK, links)
}

func createLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var input struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := decoder.Decode(&input); err != nil {
		writeError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.URL) == "" {
		writeError(w, "title and url are required", http.StatusBadRequest)
		return
	}
	link, err := store.Create(context.Background(), input.Title, input.URL)
	if err != nil {
		writeError(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

func deleteLinkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	idText := strings.TrimPrefix(r.URL.Path, "/api/links/")
	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		writeError(w, "link not found", http.StatusNotFound)
		return
	}
	if err := store.Delete(context.Background(), int64(id)); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, "link not found", http.StatusNotFound)
			return
		}
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, webFile("index.html"))
}

func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	raw := r.URL.RequestURI()
	if strings.Contains(raw, "..") || strings.Contains(r.URL.Path, "..") {
		http.NotFound(w, r)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/static/")
	if name == "" || filepath.IsAbs(name) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, webFile(filepath.Clean(name)))
}

func webFile(name string) string {
	if root, err := os.Getwd(); err == nil {
		for dir := root; ; dir = filepath.Dir(dir) {
			candidate := filepath.Join(dir, "web", name)
			if _, err := os.Stat(candidate); err == nil {
				return candidate
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
		}
	}
	return filepath.Join("web", name)
}

func writeJSON(w http.ResponseWriter, status int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, status, map[string]string{"error": message})
}
