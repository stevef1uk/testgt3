package api

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == "/" && r.Method == http.MethodGet:
		h.serveRoot(w, r)
	case strings.HasPrefix(r.URL.Path, "/static/") && r.Method == http.MethodGet:
		h.serveStatic(w, r)
	case r.URL.Path == "/api/links" && r.Method == http.MethodGet:
		h.listLinks(w, r)
	case r.URL.Path == "/api/links" && r.Method == http.MethodPost:
		h.createLink(w, r)
	case strings.HasPrefix(r.URL.Path, "/api/links/") && r.Method == http.MethodDelete:
		h.deleteLink(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) serveRoot(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "index.html"))
}

func (h *Handler) serveStatic(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/static/")
	if name == "" || strings.Contains(name, "..") {
		http.Error(w, "invalid static path", http.StatusBadRequest)
		return
	}
	clean := filepath.Clean(name)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		http.Error(w, "invalid static path", http.StatusBadRequest)
		return
	}
	http.ServeFile(w, r, filepath.Join("web", clean))
}

func (h *Handler) listLinks(w http.ResponseWriter, r *http.Request) {
	links, err := store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, links)
}

func (h *Handler) createLink(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	link, err := store.Create(r.Context(), input.Title, input.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

func (h *Handler) deleteLink(w http.ResponseWriter, r *http.Request) {
	rawID := strings.TrimPrefix(r.URL.Path, "/api/links/")
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid link id")
		return
	}
	if err := store.Delete(r.Context(), id); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		writeError(w, status, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
