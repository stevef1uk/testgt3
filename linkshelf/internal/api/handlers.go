package api

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

type handler struct{}

func NewPersistentHandler(_ ...any) http.Handler {
	return handler{}
}

func NewHandler(args ...any) http.Handler {
	return NewPersistentHandler(args...)
}

func (handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/links":
		links, err := store.List(context.Background())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, links)
	case r.Method == http.MethodPost && r.URL.Path == "/api/links":
		var input struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil ||
			strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.URL) == "" {
			writeError(w, http.StatusBadRequest, "invalid link")
			return
		}
		link, err := store.Create(context.Background(), input.Title, input.URL)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, link)
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/links/"):
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/links/"))
		if err != nil {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}
		if err := store.Delete(context.Background(), int64(id)); err != nil {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.NotFound(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.RequestURI(), "..") {
		http.NotFound(w, r)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/static/")
	if name == "" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join("web", filepath.Clean(name)))
}
