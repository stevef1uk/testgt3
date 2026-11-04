package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"linkshelf/internal/api"
)

func TestServerHandlerIsReachable(t *testing.T) {
	handler := api.NewHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/items status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got == "" {
		t.Fatal("GET /api/items did not set Content-Type")
	}
}
