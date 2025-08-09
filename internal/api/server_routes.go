package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewServerRouter sets up chi routes for server endpoints
func NewServerRouter(h *ServerHandlers) http.Handler {
	r := chi.NewRouter()

	r.Post("/server", h.ProvisionServer)
	r.Get("/servers/{id}", h.GetServer)
	r.Post("/servers/{id}/action", h.ServerAction)
	r.Get("/servers", h.ListServers)
	r.Get("/servers/{id}/logs", h.GetServerLogs)

	return r
}
