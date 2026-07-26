package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"

	"linkshelf/internal/store"
)

// RegisterHandlers registers all HTTP handlers on the default serve mux.
func RegisterHandlers() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/static/", handleStatic)
	http.HandleFunc("/api/links", handleLinks)       // GET and POST
	http.HandleFunc("/api/links/", handleLinkDelete) // DELETE /api/links/{id}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, path.Join("web", "index.html"))
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Reject paths with ".."
	if strings.Contains(r.URL.Path, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	filePath := strings.TrimPrefix(r.URL.Path, "/static/")
	fullPath := path.Join("web", filePath)
	// Prevent directory traversal: ensure fullPath stays under "web"
	if !strings.HasPrefix(fullPath, "web") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	http.ServeFile(w, r, fullPath)
}

func handleLinks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listLinks(w, r)
	case http.MethodPost:
		createLink(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleLinkDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Extract ID from path: /api/links/{id}
	idStr := strings.TrimPrefix(r.URL.Path, "/api/links/")
	if idStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing link ID"})
		return
	}
	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid link ID"})
		return
	}
	if err := store.Delete(r.Context(), id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func listLinks(w http.ResponseWriter, r *http.Request) {
	links, err := store.List(r.Context())
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(links)
}

func createLink(w http.ResponseWriter, r *http.Request) {
	var link store.Link
	if err := json.NewDecoder(r.Body).Decode(&link); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}
	if link.Title == "" || link.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Title and URL are required"})
		return
	}
	created, err := store.Create(r.Context(), link.Title, link.URL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}
