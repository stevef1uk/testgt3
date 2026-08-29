package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

// newTestServer brings up an httptest.Server with a fresh in-memory
// SQLite database and the package handlers registered on a local
// mux. It returns the base URL of the server.
//
// The webRoot is overridden to a temp dir so static / index tests
// have a known filesystem to work against. The previous webRoot is
// restored on cleanup.
func newTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.InitSchema(db); err != nil {
		db.Close()
		t.Fatal(err)
	}
	store.DB = db
	t.Cleanup(func() {
		db.Close()
		store.DB = nil
	})

	tmp := t.TempDir()
	prevWeb := webRoot
	webRoot = tmp
	t.Cleanup(func() { webRoot = prevWeb })

	mux := http.NewServeMux()
	RegisterRoutes(mux)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, srv.URL
}

func decodeLinks(t *testing.T, r io.Reader) []store.Link {
	t.Helper()
	var links []store.Link
	if err := json.NewDecoder(r).Decode(&links); err != nil {
		t.Fatalf("decode links: %v", err)
	}
	return links
}

// TestListEmptyReturnsEmptyArray verifies that GET /api/links with no
// rows in the store returns 200 with a JSON array (not null) so the
// client can iterate without a nil check.
func TestListEmptyReturnsEmptyArray(t *testing.T) {
	_, base := newTestServer(t)

	resp, err := http.Get(base + "/api/links")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json...", ct)
	}
	body, _ := io.ReadAll(resp.Body)
	if got := strings.TrimSpace(string(body)); got != "[]" {
		t.Fatalf("body = %q, want []", got)
	}
}

// TestCreateAndListHappyPath verifies the POST/GET cycle.
func TestCreateAndListHappyPath(t *testing.T) {
	_, base := newTestServer(t)

	body := bytes.NewBufferString(`{"title":"Example","url":"https://example.com"}`)
	resp, err := http.Post(base+"/api/links", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		raw, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, want 201; body = %s", resp.StatusCode, raw)
	}

	var created store.Link
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == 0 || created.Title != "Example" || created.URL != "https://example.com" {
		t.Fatalf("created = %+v, want populated link", created)
	}
	if created.CreatedAt == "" {
		t.Fatal("created.CreatedAt is empty")
	}

	resp2, err := http.Get(base + "/api/links")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	links := decodeLinks(t, resp2.Body)
	if len(links) != 1 || links[0].ID != created.ID {
		t.Fatalf("list = %+v, want one matching link", links)
	}
}

// TestCreateValidationErrorReturns400 verifies the 400 error contract.
func TestCreateValidationErrorReturns400(t *testing.T) {
	_, base := newTestServer(t)

	cases := []struct {
		name string
		body string
	}{
		{"empty title", `{"title":"","url":"https://example.com"}`},
		{"empty url", `{"title":"x","url":""}`},
		{"bad scheme", `{"title":"x","url":"ftp://example.com"}`},
		{"bad json", `not json`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := http.Post(base+"/api/links", "application/json", bytes.NewBufferString(tc.body))
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusBadRequest {
				raw, _ := io.ReadAll(resp.Body)
				t.Fatalf("status = %d, want 400; body = %s", resp.StatusCode, raw)
			}
			var errBody map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
				t.Fatalf("decode error body: %v", err)
			}
			if errBody["error"] == "" {
				t.Fatal("error body missing 'error' field")
			}
		})
	}
}

// TestDeleteHappyPath verifies a 204 on success.
func TestDeleteHappyPath(t *testing.T) {
	_, base := newTestServer(t)

	body := bytes.NewBufferString(`{"title":"x","url":"https://example.com"}`)
	resp, err := http.Post(base+"/api/links", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	var created store.Link
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	req, err := http.NewRequest(http.MethodDelete, base+"/api/links/"+strconvFormat(created.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	delResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer delResp.Body.Close()
	if delResp.StatusCode != http.StatusNoContent {
		raw, _ := io.ReadAll(delResp.Body)
		t.Fatalf("status = %d, want 204; body = %s", delResp.StatusCode, raw)
	}
}

// TestDeleteMissingReturns404 verifies the 404 error contract.
func TestDeleteMissingReturns404(t *testing.T) {
	_, base := newTestServer(t)

	req, err := http.NewRequest(http.MethodDelete, base+"/api/links/99999", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
	var errBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
		t.Fatalf("decode error body: %v", err)
	}
	if errBody["error"] == "" {
		t.Fatal("error body missing 'error' field")
	}
}

// TestStaticRejectsDotDot verifies path-traversal protection.
func TestStaticRejectsDotDot(t *testing.T) {
	_, base := newTestServer(t)

	cases := []string{
		"/static/../etc/passwd",
		"/static/foo/../../bar",
	}
	for _, p := range cases {
		t.Run(p, func(t *testing.T) {
			resp, err := http.Get(base + p)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", resp.StatusCode)
			}
		})
	}
}

// TestStaticServesFile verifies happy-path file serving.
func TestStaticServesFile(t *testing.T) {
	_, base := newTestServer(t)

	tmp := t.TempDir()
	want := "hello world"
	if err := os.WriteFile(filepath.Join(tmp, "hello.txt"), []byte(want), 0o644); err != nil {
		t.Fatal(err)
	}
	prev := webRoot
	webRoot = tmp
	t.Cleanup(func() { webRoot = prev })

	resp, err := http.Get(base + "/static/hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != want {
		t.Fatalf("body = %q, want %q", body, want)
	}
}

// TestIndexServesIndexHTML verifies the / route serves index.html.
func TestIndexServesIndexHTML(t *testing.T) {
	_, base := newTestServer(t)

	tmp := t.TempDir()
	html := "<html><body>hi</body></html>"
	if err := os.WriteFile(filepath.Join(tmp, "index.html"), []byte(html), 0o644); err != nil {
		t.Fatal(err)
	}
	prev := webRoot
	webRoot = tmp
	t.Cleanup(func() { webRoot = prev })

	resp, err := http.Get(base + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != html {
		t.Fatalf("body = %q, want %q", body, html)
	}
}

// strconvFormat formats an int64 as base-10. We don't pull in strconv
// at the top of the file just for this; keep the helper local.
func strconvFormat(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
