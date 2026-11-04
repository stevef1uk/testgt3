package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"linkshelf/internal/store"
)

type Handler struct {
	mu    sync.RWMutex
	links map[int64]store.Link
	next  int64
}

func NewHandler() *Handler {
	return &Handler{links: make(map[int64]store.Link)}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/links" || r.URL.Path == "/api/items" {
		switch r.Method {
		case http.MethodGet:
			h.list(w)
		case http.MethodPost:
			h.create(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	var prefix string
	switch {
	case strings.HasPrefix(r.URL.Path, "/api/links/"):
		prefix = "/api/links/"
	case strings.HasPrefix(r.URL.Path, "/api/item/"):
		prefix = "/api/item/"
	default:
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := strconv.ParseInt(strings.TrimPrefix(r.URL.Path, prefix), 10, 64)
	if err != nil || id <= 0 {
		http.NotFound(w, r)
		return
	}
	h.mu.Lock()
	_, exists := h.links[id]
	if exists {
		delete(h.links, id)
	}
	h.mu.Unlock()
	if !exists {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) list(w http.ResponseWriter) {
	h.mu.RLock()
	links := make([]store.Link, 0, len(h.links))
	for _, link := range h.links {
		links = append(links, link)
	}
	h.mu.RUnlock()
	writeJSON(w, http.StatusOK, links)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.URL = strings.TrimSpace(input.URL)
	if input.Title == "" || input.URL == "" {
		http.Error(w, "title and url are required", http.StatusBadRequest)
		return
	}

	h.mu.Lock()
	h.next++
	link := store.Link{
		ID:        h.next,
		Title:     input.Title,
		URL:       input.URL,
		CreatedAt: time.Now().UTC().Truncate(time.Second),
	}
	h.links[link.ID] = link
	h.mu.Unlock()
	writeJSON(w, http.StatusCreated, link)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
