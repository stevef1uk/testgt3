// Package api contains the HTTP layer for Link Shelf.
//
// Handlers are registered on an *http.ServeMux by RegisterRoutes. All
// persistence is delegated to the store package; this file owns no SQL.
package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"linkshelf/internal/store"
)

// webRoot is the directory that index.html and static assets live in.
// It is overridden in tests via setWebRoot to point at a temp dir.
var webRoot = "web"

// setWebRoot overrides the directory used by indexHandler and
// staticFileHandler. It is intended for use in tests only.
func setWebRoot(path string) {
	webRoot = path
}

// RegisterRoutes attaches every Link Shelf HTTP route to the supplied
// mux. Callers are expected to pass http.DefaultServeMux.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/static/", staticFileHandler)
	mux.HandleFunc("/api/links", linksCollectionHandler)
	mux.HandleFunc("/api/links/", linkItemHandler)
}

// linksCollectionHandler dispatches /api/links to the appropriate
// handler based on HTTP method.
func linksCollectionHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		listLinksHandler(w, r)
	case http.MethodPost:
		createLinkHandler(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// linkItemHandler dispatches /api/links/{id} to the appropriate handler
// based on HTTP method.
func linkItemHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/links/")
	idStr = strings.TrimSuffix(idStr, "/")
	if idStr == "" {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusNotFound, "link not found")
		return
	}
	switch r.Method {
	case http.MethodDelete:
		deleteLinkHandler(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// listLinksHandler responds with a JSON array of all links. When the
// store is empty the response is [] (not null) so client code can
// iterate without a nil check.
func listLinksHandler(w http.ResponseWriter, r *http.Request) {
	links, err := store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if links == nil {
		links = []store.Link{}
	}
	writeJSON(w, http.StatusOK, links)
}

// createLinkHandler accepts a JSON body of the form
//
//	{"title": "...", "url": "..."}
//
// and creates a new link. On success it returns 201 with the created
// link as JSON. Validation errors from store.Create are returned as
// 400 with an error JSON body.
func createLinkHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	link, err := store.Create(r.Context(), body.Title, body.URL)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

// deleteLinkHandler removes a link by id. On success it returns 204
// with no body. Missing rows produce a 404.
func deleteLinkHandler(w http.ResponseWriter, r *http.Request, id int64) {
	if err := store.Delete(r.Context(), id); err != nil {
		if errors.Is(err, errLinkNotFound) || err.Error() == "sql: no rows in result set" {
			writeError(w, http.StatusNotFound, "link not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// errLinkNotFound is returned by store.Delete when no row was deleted.
// We declare it here so handlers can use errors.Is without depending
// on database/sql. The store package surfaces sql.ErrNoRows for this
// case; we re-wrap it via a sentinel for clarity.
var errLinkNotFound = errors.New("link not found")

// indexHandler serves web/index.html. If the file is missing it
// returns 404 with a plain text body.
func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	path := filepath.Join(webRoot, "index.html")
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}

// staticFileHandler serves a file from web/{trimmed path}, where the
// trimmed path is the URL path with the "/static/" prefix removed.
// Any path containing ".." is rejected with 404 to keep the handler
// inside the web directory.
func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	rel := strings.TrimPrefix(r.URL.Path, "/static/")
	if rel == "" || strings.Contains(rel, "..") {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(webRoot, rel)
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}

// writeJSON writes v as JSON with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error body {"error": "..."} with the
// given status code.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
