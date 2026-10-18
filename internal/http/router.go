package apphttp

import (
	"net/http"

	"github.com/AlbertSalimov/streamhub/internal/content"
	"github.com/AlbertSalimov/streamhub/internal/http/handler"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(db *pgxpool.Pool, deps ...any) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", handler.Health)

	readiness := &handler.ReadinessHandler{
		DB: db,
	}
	r.Get("/readyz", readiness.ServeHTTP)

	repository := content.NewRepository(db)
	service := content.NewService(repository)

	for _, dep := range deps {
		switch value := dep.(type) {
		case content.Cache:
			service.SetCache(value)
		case content.EventPublisher:
			service.SetPublisher(value)
		}
	}
	contentHandler := handler.NewContentHandler(service)

	r.Route("/api/v1/content", func(r chi.Router) {
		r.Post("/", contentHandler.Create)
		r.Get("/", contentHandler.List)
		r.Get("/{id}", contentHandler.GetByID)
		r.Patch("/{id}", contentHandler.Update)
		r.Delete("/{id}", contentHandler.Delete)
	})

	return r
}
