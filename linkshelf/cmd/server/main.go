// Command server runs the Link Shelf HTTP service.
package main

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"linkshelf/internal/api"
	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func registerAPI(mux *http.ServeMux) {
	api.RegisterRoutes(mux)
}

func serveStaticFiles(mux *http.ServeMux) {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "web/index.html")
	})
	fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir("web")))
	mux.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "..") {
			http.NotFound(w, r)
			return
		}
		fileServer.ServeHTTP(w, r)
	}))
}

func main() {
	db, err := sql.Open("sqlite3", "linkshelf.db")
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		log.Fatalf("init schema: %v", err)
	}
	store.DB = db

	registerAPI(http.DefaultServeMux)
	serveStaticFiles(http.DefaultServeMux)

	log.Println("listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
