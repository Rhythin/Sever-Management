package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rhythin/sever-management/internal/handlers"
)

// NewServerRouter sets up chi routes for server endpoints
func NewServerRouter(h handlers.ServerHandler) http.Handler {
	r := chi.NewRouter()

	r.Post("/server", h.ProvisionServer)

	r.Route("/servers", func(r chi.Router) {
		r.Get("/", h.ListServers)
		r.Get("/{id}", h.GetServer)
		r.Post("/{id}/action", h.ServerAction)
		r.Get("/{id}/logs", h.GetServerLogs)
	})

	return r
}
