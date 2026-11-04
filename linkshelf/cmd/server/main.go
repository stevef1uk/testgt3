package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"linkshelf/internal/api"
	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func newServer() http.Handler {
	handler := api.NewHandler()
	http.Handle("/api/", handler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" && r.Method == http.MethodGet {
			http.ServeFile(w, r, filepath.Join("web", "index.html"))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/static/") &&
			r.Method == http.MethodGet &&
			!strings.Contains(r.URL.Path, "..") {
			http.ServeFile(w, r, filepath.Join("web", strings.TrimPrefix(r.URL.Path, "/static/")))
			return
		}
		http.NotFound(w, r)
	})

	return http.DefaultServeMux
}

func main() {
	db, err := sql.Open("sqlite3", "linkshelf.db")
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		log.Fatalf("initialize schema: %v", err)
	}
	store.DB = db

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, newServer()); err != nil {
		log.Fatalf("serve HTTP: %v", err)
	}
}
