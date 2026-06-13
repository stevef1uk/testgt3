package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/example/linkshelf/internal/api"
	"github.com/example/linkshelf/internal/store"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Open SQLite database (file‑based for persistence)
	db, err := sql.Open("sqlite3", "file:linkshelf.db?cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Initialise the schema (creates the links table if missing)
	if err := store.InitSchema(db); err != nil {
		log.Fatalf("failed to initialise schema: %v", err)
	}

	// Expose the DB to the package‑level variable used by store functions
	store.DB = db

	// HTTP mux for routing
	mux := http.NewServeMux()

	// UI routes
	mux.HandleFunc("/", api.HandleRoot)
	mux.HandleFunc("/static/", api.HandleStatic)

	// API routes – composite handler for /api/links (GET list, POST create)
	mux.HandleFunc("/api/links", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			api.HandleList(w, r)
		case http.MethodPost:
			api.HandleCreate(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// DELETE endpoint for /api/links/{id}
	mux.HandleFunc("/api/links/", api.HandleDelete)

	// Start HTTP server
	addr := ":8080"
	log.Printf("starting server on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
