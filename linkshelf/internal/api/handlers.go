package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"linkshelf/internal/store"
)

type link struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/links", links)
	mux.HandleFunc("/api/links/", linkByID)
}

func links(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list(w, r)
	case http.MethodPost:
		create(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func linkByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/api/links/"))
	if err != nil || id < 1 {
		http.Error(w, "invalid link id", http.StatusBadRequest)
		return
	}
	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", "DELETE")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	result, err := store.DB.ExecContext(r.Context(), `DELETE FROM links WHERE id = ?`, id)
	if err != nil {
		http.Error(w, "delete link: "+err.Error(), http.StatusInternalServerError)
		return
	}
	count, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "delete link: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if count == 0 {
		http.NotFound(w, r)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func list(w http.ResponseWriter, r *http.Request) {
	rows, err := store.DB.QueryContext(r.Context(),
		`SELECT id, title, url, created_at FROM links ORDER BY id`)
	if err != nil {
		http.Error(w, "list links: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	result := make([]link, 0)
	for rows.Next() {
		var item link
		if err := rows.Scan(&item.ID, &item.Title, &item.URL, &item.CreatedAt); err != nil {
			http.Error(w, "list links: "+err.Error(), http.StatusInternalServerError)
			return
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "list links: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	input.URL = strings.TrimSpace(input.URL)
	if input.Title == "" || input.URL == "" {
		http.Error(w, "title and url are required", http.StatusBadRequest)
		return
	}

	result, err := store.DB.ExecContext(r.Context(),
		`INSERT INTO links (title, url) VALUES (?, ?)`, input.Title, input.URL)
	if err != nil {
		http.Error(w, "create link: "+err.Error(), http.StatusInternalServerError)
		return
	}
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "create link: "+err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, link{ID: int(id), Title: input.Title, URL: input.URL})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
