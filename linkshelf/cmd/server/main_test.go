package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterAPIRegistersRoutes(t *testing.T) {
	mux := http.NewServeMux()
	registerAPI(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/links", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code == http.StatusNotFound {
		t.Fatalf("expected API route to be registered, got %d", rec.Code)
	}
}

func TestServeStaticFilesServesRoot(t *testing.T) {
	t.Chdir("../..")

	mux := http.NewServeMux()
	serveStaticFiles(mux)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected root to be served, got %d", rec.Code)
	}
}
