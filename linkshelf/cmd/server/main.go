package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"

	"linkshelf/internal/store"
)

type linkPayload struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// handleLinks handles GET (list) and POST (create) on /api/links
func handleLinks(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		links, err := store.List(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := json.NewEncoder(w).Encode(links); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		var payload linkPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		created, err := store.Create(r.Context(), payload.Title, payload.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(created); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleLink handles DELETE on /api/links/{id}
func handleLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Expect path /api/links/{id}
	idStr := r.URL.Path[len("/api/links/"):]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := store.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func main() {
	// Initialize the database connection
	db, err := sql.Open("sqlite3", "linkshelf.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize the schema (tables)
	if err := store.InitSchema(db); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Assign the global DB variable used by store package
	store.DB = db

	// Register HTTP handlers
	http.HandleFunc("/api/links", handleLinks)
	// Handles DELETE /api/links/{id}
	http.HandleFunc("/api/links/", handleLink)

	// Serve static web assets (index.html, app.js, etc.)
	fs := http.FileServer(http.Dir("web"))
	http.Handle("/", fs)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
