package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AlbertSalimov/streamhub/internal/http/handler"
)

func NewRouter(db *pgxpool.Pool) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)

	router.Get("/healthz", handler.Health)

	readiness := &handler.ReadinessHandler{
		DB: db,
	}
	router.Get("/readyz", readiness.ServeHTTP)

	return router
}
