package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCreateRejectsInvalidJSON(t *testing.T) {
	h := NewContentHandler(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/content/",
		strings.NewReader(`{"title":`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestGetByIDRejectsInvalidUUID(t *testing.T) {
	h := NewContentHandler(nil)

	r := chi.NewRouter()
	r.Get("/content/{id}", h.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/content/not-a-uuid", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}
