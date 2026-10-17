package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestReadinessReturnsUnavailableWhenDatabaseIsClosed(t *testing.T) {
	db, err := pgxpool.New(
		t.Context(),
		"postgres://user:password@localhost:5432/streamhub?sslmode=disable",
	)
	if err != nil {
		t.Fatalf("create pool: %v", err)
	}
	db.Close()

	handler := &ReadinessHandler{DB: db}

	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusServiceUnavailable,
			rec.Code,
		)
	}
}
