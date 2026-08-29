package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"linkshelf/internal/api"
	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InitSchema(db); err != nil {
		t.Fatal(err)
	}
	store.DB = db
	mux := http.NewServeMux()
	api.RegisterRoutes(mux)
	return httptest.NewServer(mux)
}

func TestLinksAPI(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/links")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET empty status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var links []store.Link
	if err := json.NewDecoder(resp.Body).Decode(&links); err != nil {
		t.Fatal(err)
	}
	if links == nil || len(links) != 0 {
		t.Fatalf("GET empty body = %#v, want []", links)
	}

	resp, err = http.Post(srv.URL+"/api/links", "application/json", bytes.NewBufferString(`{"title":"Go","url":"https://go.dev"}`))
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST status = %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var created store.Link
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if created.ID == 0 || created.Title != "Go" || created.URL != "https://go.dev" {
		t.Fatalf("created link = %#v", created)
	}

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/links/"+strconv.FormatInt(created.ID, 10), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
}

func TestLinksAPIRejectsBadInputAndMissingDelete(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/links", "application/json", bytes.NewBufferString(`{"title":"","url":""}`))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("bad POST status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}

	req, err := http.NewRequest(http.MethodDelete, srv.URL+"/api/links/999", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("missing DELETE status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestIndexAndStaticRoutes(t *testing.T) {
	srv := testServer(t)
	defer srv.Close()

	for _, path := range []string{"/", "/static/index.html"} {
		resp, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", path, resp.StatusCode, http.StatusOK)
		}
	}
	resp, err := http.Get(srv.URL + "/static/../go.mod")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("traversal status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}
