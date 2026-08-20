package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

type linkInput struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/static/", staticHandler)
	mux.HandleFunc("/api/links", linksHandler)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, webPath("index.html"))
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	if strings.Contains(r.URL.RequestURI(), "..") || strings.Contains(r.URL.RawPath, "..") {
		http.NotFound(w, r)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/static/")
	if name == "" {
		http.NotFound(w, r)
		return
	}
	clean := filepath.Clean(name)
	if clean == ".." {
		http.NotFound(w, r)
		return
	}
	prefix := ".." + string(os.PathSeparator)
	if strings.HasPrefix(clean, prefix) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, webPath(clean))
}

func linksHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/links" {
		if r.Method == http.MethodGet {
			listLinks(w, r)
			return
		}
		if r.Method == http.MethodPost {
			createLink(w, r)
			return
		}
	}
	if r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/links/") {
		deleteLink(w, r)
		return
	}
	http.NotFound(w, r)
}

func listLinks(w http.ResponseWriter, r *http.Request) {
	links, err := store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if links == nil {
		links = make([]store.Link, 0)
	}
	writeJSON(w, http.StatusOK, links)
}

func createLink(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var input linkInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, errors.New("invalid JSON"))
		return
	}
	if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.URL) == "" {
		writeError(w, http.StatusBadRequest, errors.New("title and url are required"))
		return
	}
	link, err := store.Create(r.Context(), input.Title, input.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

func deleteLink(w http.ResponseWriter, r *http.Request) {
	rawID := strings.TrimPrefix(r.URL.Path, "/api/links/")
	if rawID == "" || strings.Contains(rawID, "/") {
		writeError(w, http.StatusNotFound, errors.New("link not found"))
		return
	}
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusNotFound, errors.New("link not found"))
		return
	}
	if err := store.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func webPath(name string) string {
	wd, err := os.Getwd()
	if err != nil {
		return filepath.Join("web", name)
	}
	for dir := wd; ; dir = filepath.Dir(dir) {
		candidate := filepath.Join(dir, "web", name)
		if _, statErr := os.Stat(candidate); statErr == nil {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return filepath.Join("web", name)
}
