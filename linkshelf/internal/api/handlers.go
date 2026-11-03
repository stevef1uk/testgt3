package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/links", links)
	mux.HandleFunc("/api/links/", linkByID)
}

func links(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		items, err := store.List(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var input struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		item, err := store.Create(r.Context(), input.Title, input.URL)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, item)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func linkByID(w http.ResponseWriter, r *http.Request) {
	rawID := strings.TrimPrefix(r.URL.Path, "/api/links/")
	id, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || id < 1 {
		writeError(w, http.StatusBadRequest, "invalid link id")
		return
	}
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := store.Delete(r.Context(), id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
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
