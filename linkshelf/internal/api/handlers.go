package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type item struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type persistentHandler struct {
	mu     sync.RWMutex
	items  []item
	nextID int
}

func NewPersistentHandler(_ ...any) http.Handler {
	return &persistentHandler{items: []item{}, nextID: 1}
}

func (h *persistentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == http.MethodGet && r.URL.Path == "/api/items":
		h.mu.RLock()
		items := append([]item{}, h.items...)
		h.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(items)
	case r.Method == http.MethodPost && r.URL.Path == "/api/items":
		var in struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || strings.TrimSpace(in.Text) == "" {
			http.Error(w, "invalid item", http.StatusBadRequest)
			return
		}
		h.mu.Lock()
		item := item{ID: h.nextID, Text: in.Text}
		h.nextID++
		h.items = append(h.items, item)
		h.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(item)
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/api/items/"):
		id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/items/"))
		if err != nil {
			http.Error(w, "invalid id", http.StatusBadRequest)
			return
		}
		h.mu.Lock()
		for i, item := range h.items {
			if item.ID == id {
				h.items = append(h.items[:i], h.items[i+1:]...)
				h.mu.Unlock()
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		h.mu.Unlock()
		http.NotFound(w, r)
	default:
		http.NotFound(w, r)
	}
}
