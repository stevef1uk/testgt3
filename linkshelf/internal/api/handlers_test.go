package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"linkshelf/internal/api"
	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

type responseRecorder struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func newResponseRecorder() *responseRecorder {
	return &responseRecorder{header: make(http.Header)}
}

func (r *responseRecorder) Header() http.Header {
	return r.header
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.body.Write(body)
}

func newTestServer(t *testing.T) *http.ServeMux {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := store.InitSchema(db); err != nil {
		t.Fatalf("initialize schema: %v", err)
	}
	store.DB = db

	mux := http.NewServeMux()
	api.RegisterRoutes(mux)
	return mux
}

func performRequest(t *testing.T, mux *http.ServeMux, method, path, body string) *responseRecorder {
	t.Helper()

	req, err := http.NewRequest(method, path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := newResponseRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func responseStatus(rec *responseRecorder) int {
	if rec.status == 0 {
		return http.StatusOK
	}
	return rec.status
}

func TestLinksAPI(t *testing.T) {
	mux := newTestServer(t)

	t.Run("GET returns empty array", func(t *testing.T) {
		rec := performRequest(t, mux, http.MethodGet, "/api/links", "")
		if responseStatus(rec) != http.StatusOK {
			t.Fatalf("status = %d, want %d", responseStatus(rec), http.StatusOK)
		}

		var links []store.Link
		if err := json.Unmarshal(rec.body.Bytes(), &links); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if links == nil || len(links) != 0 {
			t.Fatalf("links = %#v, want a non-nil empty array", links)
		}
	})

	t.Run("POST creates and GET lists a link", func(t *testing.T) {
		rec := performRequest(t, mux, http.MethodPost, "/api/links",
			`{"title":"Go","url":"https://go.dev"}`)
		if responseStatus(rec) != http.StatusCreated {
			t.Fatalf("POST status = %d, want %d", responseStatus(rec), http.StatusCreated)
		}

		var created store.Link
		if err := json.Unmarshal(rec.body.Bytes(), &created); err != nil {
			t.Fatalf("decode created link: %v", err)
		}
		if created.ID == 0 || created.Title != "Go" || created.URL != "https://go.dev" {
			t.Fatalf("created link = %#v", created)
		}

		rec = performRequest(t, mux, http.MethodGet, "/api/links", "")
		var links []store.Link
		if err := json.Unmarshal(rec.body.Bytes(), &links); err != nil {
			t.Fatalf("decode links: %v", err)
		}
		if len(links) != 1 || links[0].ID != created.ID {
			t.Fatalf("links = %#v, want created link", links)
		}
	})

	t.Run("invalid POST returns bad request", func(t *testing.T) {
		for name, body := range map[string]string{
			"malformed JSON": `{"title":`,
			"invalid fields": `{"title":"","url":"ftp://example.com"}`,
		} {
			t.Run(name, func(t *testing.T) {
				rec := performRequest(t, mux, http.MethodPost, "/api/links", body)
				if responseStatus(rec) != http.StatusBadRequest {
					t.Fatalf("status = %d, want %d", responseStatus(rec), http.StatusBadRequest)
				}
			})
		}
	})

	t.Run("DELETE removes a link and reports missing IDs", func(t *testing.T) {
		rec := performRequest(t, mux, http.MethodPost, "/api/links",
			`{"title":"Delete me","url":"https://example.com"}`)

		var created store.Link
		if err := json.Unmarshal(rec.body.Bytes(), &created); err != nil {
			t.Fatalf("decode created link: %v", err)
		}

		path := "/api/links/" + strconv.FormatInt(created.ID, 10)
		rec = performRequest(t, mux, http.MethodDelete, path, "")
		if responseStatus(rec) != http.StatusNoContent {
			t.Fatalf("DELETE status = %d, want %d", responseStatus(rec), http.StatusNoContent)
		}

		rec = performRequest(t, mux, http.MethodDelete, path, "")
		if responseStatus(rec) != http.StatusNotFound {
			t.Fatalf("missing DELETE status = %d, want %d", responseStatus(rec), http.StatusNotFound)
		}
	})
}
