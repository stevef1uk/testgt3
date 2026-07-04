package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

// RegisterHandlers sets up HTTP routes on the default serve mux.
func RegisterHandlers() {
	http.HandleFunc("/api/links", handleLinks)
	http.HandleFunc("/api/links/", handleDeleteLink)
	http.HandleFunc("/static/", handleStatic)
	http.HandleFunc("/", handleRoot)
}

func handleLinks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		links, err := store.List(r.Context())
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(links)

	case http.MethodPost:
		var req struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.Title == "" || req.URL == "" {
			writeJSONError(w, "title and url are required", http.StatusBadRequest)
			return
		}
		link, err := store.Create(r.Context(), req.Title, req.URL)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				writeJSONError(w, "link already exists", http.StatusBadRequest)
				return
			}
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(link)

	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func handleDeleteLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from /api/links/{id}
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		writeJSONError(w, "missing id", http.StatusBadRequest)
		return
	}
	idStr := parts[len(parts)-1]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = store.Delete(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			writeJSONError(w, "link not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func handleStatic(w http.ResponseWriter, r *http.Request) {
	// Reject paths containing ".."
	if strings.Contains(r.URL.RequestURI(), "..") {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Determine web root: discover from executable's working directory or fallback
	webRoot := findWebRoot()
	filePath := filepath.Join(webRoot, strings.TrimPrefix(r.URL.Path, "/static/"))

	// Security: ensure the resolved path is within webRoot
	absFile, err := filepath.Abs(filePath)
	if err != nil || !strings.HasPrefix(absFile, webRoot) {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	webRoot := findWebRoot()
	indexPath := filepath.Join(webRoot, "index.html")
	http.ServeFile(w, r, indexPath)
}

func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// findWebRoot locates the web/ directory from the working directory upwards.
func findWebRoot() string {
	wd, _ := os.Getwd()
	for {
		candidate := filepath.Join(wd, "web")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	// Fallback to current dir/web
	return filepath.Join(".", "web")
}
