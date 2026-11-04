package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type testResponseRecorder struct {
	header http.Header
	code   int
	body   bytes.Buffer
}

func newTestResponseRecorder() *testResponseRecorder {
	return &testResponseRecorder{header: make(http.Header)}
}

func (r *testResponseRecorder) Header() http.Header {
	return r.header
}

func (r *testResponseRecorder) WriteHeader(code int) {
	r.code = code
}

func (r *testResponseRecorder) Write(p []byte) (int, error) {
	if r.code == 0 {
		r.code = http.StatusOK
	}
	return r.body.Write(p)
}

func makeRequest(t *testing.T, method, path, body string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, path, bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	return req
}

func TestLinksCollectionCRUD(t *testing.T) {
	h := NewHandler()

	t.Run("empty collection is JSON array", func(t *testing.T) {
		rec := newTestResponseRecorder()
		h.ServeHTTP(rec, makeRequest(t, http.MethodGet, "/api/links", ""))

		if rec.code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.code, http.StatusOK)
		}
		var links []Link
		if err := json.Unmarshal(rec.body.Bytes(), &links); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if links == nil || len(links) != 0 {
			t.Fatalf("links = %#v, want empty array", links)
		}
	})

	t.Run("create, list, and delete", func(t *testing.T) {
		rec := newTestResponseRecorder()
		h.ServeHTTP(rec, makeRequest(t, http.MethodPost, "/api/links",
			`{"title":"Go","url":"https://go.dev"}`))

		if rec.code != http.StatusCreated {
			t.Fatalf("create status = %d, want %d", rec.code, http.StatusCreated)
		}
		var created Link
		if err := json.Unmarshal(rec.body.Bytes(), &created); err != nil {
			t.Fatalf("decode created link: %v", err)
		}
		if created.ID == "" || created.Title != "Go" || created.URL != "https://go.dev" {
			t.Fatalf("created link = %#v", created)
		}

		rec = newTestResponseRecorder()
		h.ServeHTTP(rec, makeRequest(t, http.MethodGet, "/api/links", ""))
		var links []Link
		if err := json.Unmarshal(rec.body.Bytes(), &links); err != nil {
			t.Fatalf("decode list: %v", err)
		}
		if len(links) != 1 || links[0] != created {
			t.Fatalf("links = %#v, want %#v", links, []Link{created})
		}

		rec = newTestResponseRecorder()
		h.ServeHTTP(rec, makeRequest(t, http.MethodDelete, "/api/links/"+created.ID, ""))
		if rec.code != http.StatusNoContent {
			t.Fatalf("delete status = %d, want %d", rec.code, http.StatusNoContent)
		}
	})
}

func TestLinksRejectInvalidRequests(t *testing.T) {
	h := NewHandler()
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		want   int
	}{
		{"missing title", http.MethodPost, "/api/links", `{"url":"https://go.dev"}`, http.StatusBadRequest},
		{"invalid JSON", http.MethodPost, "/api/links", `{`, http.StatusBadRequest},
		{"unsupported method", http.MethodPut, "/api/links", "", http.StatusMethodNotAllowed},
		{"missing item", http.MethodDelete, "/api/links/99", "", http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := newTestResponseRecorder()
			h.ServeHTTP(rec, makeRequest(t, tt.method, tt.path, tt.body))
			if rec.code != tt.want {
				t.Fatalf("status = %d, want %d", rec.code, tt.want)
			}
		})
	}
}

var _ io.Writer = (*testResponseRecorder)(nil)
