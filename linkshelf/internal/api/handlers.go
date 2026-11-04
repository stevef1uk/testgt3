package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Link struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

type Handler struct {
	mu    sync.RWMutex
	links []Link
	next  uint64
}

func NewHandler() http.Handler {
	return &Handler{links: []Link{}}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/" {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, "<!doctype html><html><head><title>Linkshelf</title></head><body><h1>Linkshelf</h1></body></html>")
		return
	}
	if strings.HasPrefix(path, "/static/") {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		switch strings.TrimPrefix(path, "/static/") {
		case "app.js":
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			_, _ = fmt.Fprint(w, "document.addEventListener('DOMContentLoaded', function () {});")
		default:
			http.NotFound(w, r)
		}
		return
	}
	if path == "/api/links" {
		h.collection(w, r)
		return
	}
	if strings.HasPrefix(path, "/api/links/") {
		h.item(w, r, strings.TrimPrefix(path, "/api/links/"))
		return
	}
	http.NotFound(w, r)
}

func (h *Handler) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.mu.RLock()
		links := append([]Link{}, h.links...)
		h.mu.RUnlock()
		writeJSON(w, http.StatusOK, links)
	case http.MethodPost:
		var input struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		}
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&input); err != nil || strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.URL) == "" {
			http.Error(w, "invalid link", http.StatusBadRequest)
			return
		}
		var extra any
		if err := dec.Decode(&extra); err == nil {
			http.Error(w, "invalid link", http.StatusBadRequest)
			return
		}
		h.mu.Lock()
		h.next++
		link := Link{ID: strconv.FormatUint(h.next, 10), Title: input.Title, URL: input.URL}
		h.links = append(h.links, link)
		h.mu.Unlock()
		writeJSON(w, http.StatusCreated, link)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *Handler) item(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" || strings.Contains(id, "/") || strings.Contains(id, "..") {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, link := range h.links {
		if link.ID == id {
			h.links = append(h.links[:i], h.links[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	http.NotFound(w, r)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
