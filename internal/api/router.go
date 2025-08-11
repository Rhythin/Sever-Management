package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rhythin/sever-management/internal/handlers"
	"github.com/rhythin/sever-management/internal/logging"
	"github.com/rhythin/sever-management/internal/metrics"
)

func NewRouter(serverHandler handlers.ServerHandler) http.Handler {
	r := chi.NewRouter()

	r.Use(logging.RequestIDMiddleware)

	// Health and readiness
	r.Get("/healthz", handlers.HealthzHandler)
	r.Get("/readyz", handlers.ReadyzHandler)

	// Metrics
	r.Get("/metrics", http.HandlerFunc(metrics.NewMetricsHandler().ServeHTTP))

	// Swagger UI
	r.Get("/swagger/*", handlers.NewSwaggerHandler().ServeHTTP)

	// Server API
	r.Mount("/", NewServerRouter(serverHandler))

	return r
}
