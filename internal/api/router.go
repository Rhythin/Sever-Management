package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rhythin/sever-management/internal/logging"
	"github.com/rhythin/sever-management/internal/metrics"
	"go.uber.org/zap"
)

func NewRouter(serverHandlers *ServerHandlers) http.Handler {
	r := chi.NewRouter()

	r.Use(logging.RequestIDMiddleware)

	zap.S().Info("Setting up main router and endpoints")

	// Health and readiness
	zap.S().Info("Registering /healthz and /readyz endpoints")
	r.Get("/healthz", HealthzHandler)
	r.Get("/readyz", ReadyzHandler)

	// Metrics
	zap.S().Info("Registering /metrics endpoint")
	r.Get("/metrics", http.HandlerFunc(metrics.NewMetricsHandler().ServeHTTP))

	// Swagger UI
	zap.S().Info("Registering /swagger/* endpoint")
	r.Get("/swagger/*", NewSwaggerHandler().ServeHTTP)

	// Server API
	zap.S().Info("Registering server API endpoints")
	r.Mount("/", NewServerRouter(serverHandlers))

	return r
}
