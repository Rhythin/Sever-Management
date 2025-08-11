package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rhythin/sever-management/internal/domain"
	"github.com/rhythin/sever-management/internal/logging"
	"github.com/rhythin/sever-management/internal/packets"
	"github.com/rhythin/sever-management/internal/service"
)

// ServerHandlers provides HTTP handlers for server endpoints
type serverHandler struct {
	Service service.ServerService
}

// @Summary Provision a new virtual server
// @Description Provision a new virtual server
// @Tags servers
// @Accept json
// @Produce json
// @Param server body ProvisionRequest true "Server spec"
// @Success 201 {object} map[string]string
// @Failure 400 {object} errorResponse
// @Router /server [post]
func (h *serverHandler) ProvisionServer(w http.ResponseWriter, r *http.Request) {
	log := logging.S(r.Context())
	log.Infow("POST /server - ProvisionServer called")
	var req packets.ProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warnw("Invalid request body", "error", err)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Region == "" || req.Type == "" {
		respondError(w, http.StatusBadRequest, "region and type are required")
		return
	}
	id, err := h.Service.Provision(r.Context(), req.Region, req.Type)
	if err != nil {
		log.Errorw("Failed to provision server", "error", err)
		if err.Error() == "no available IPs" {
			respondError(w, http.StatusConflict, "no available IPs")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to provision server")
		return
	}
	log.Infow("Provisioned server", "id", id, "request", req)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// @Summary Get server metadata
// @Description Retrieve full metadata for a server
// @Tags servers
// @Produce json
// @Param id path string true "Server ID"
// @Success 200 {object} ServerResponse
// @Failure 404 {object} errorResponse
// @Router /servers/{id} [get]
func (h *serverHandler) GetServer(w http.ResponseWriter, r *http.Request) {
	log := logging.S(r.Context())
	id := chi.URLParam(r, "id")
	log.Infow("GET /servers/{id} - GetServer called", "id", id)

	server, err := h.Service.GetServerByID(r.Context(), id)
	if err != nil {
		log.Warnw("Server not found", "id", id, "error", err)
		respondError(w, http.StatusNotFound, "server not found")
		return
	}

	// Convert to response type
	resp := packets.ServerResponse{
		ID:     server.ID,
		Region: server.Region,
		Type:   server.Type,
		State:  server.State,
	}

	// Add IP address if available
	if server.IP != nil {
		resp.IPAddress = server.IP.Address
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Errorw("Failed to encode response", "error", err)
	}
}

// @Summary Perform server action
// @Description Perform an action (start, stop, reboot, terminate) on a server
// @Tags servers
// @Accept json
// @Produce json
// @Param id path string true "Server ID"
// @Param action body ActionRequest true "Action"
// @Success 200 {object} ActionResponse
// @Failure 409 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /servers/{id}/action [post]
func (h *serverHandler) ServerAction(w http.ResponseWriter, r *http.Request) {
	log := logging.S(r.Context())
	log.Infow("POST /servers/{id}/action - ServerAction called")
	id := chi.URLParam(r, "id")
	if id == "" {
		log.Warnw("Missing server ID")
		respondError(w, http.StatusBadRequest, "server ID is required")
		return
	}

	var req packets.ActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warnw("Invalid request body", "error", err)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Convert string action to domain.ServerAction and validate
	action := domain.ServerAction(req.Action)
	if !domain.IsValidAction(action) {
		log.Warnw("Invalid action", "action", req.Action)
		respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid action: %s, must be one of: start, stop, reboot, terminate", req.Action))
		return
	}

	if err := h.Service.Action(r.Context(), id, action); err != nil {
		log.Errorw("Failed to perform action", "error", err, "server_id", id, "action", action)
		if err.Error() == "invalid state transition" { // Check error message instead of type
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to perform action")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(packets.ActionResponse{
		Result: fmt.Sprintf("Action '%s' initiated successfully", action),
	})
}

// @Summary List servers
// @Description List all servers with optional filtering
// @Tags servers
// @Produce json
// @Param region query string false "Filter by region"
// @Param type query string false "Filter by server type"
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit number of results" default(20)
// @Param offset query int false "Offset for pagination" default(0)
// @Success 200 {array} ServerResponse
// @Router /servers [get]
func (h *serverHandler) ListServers(w http.ResponseWriter, r *http.Request) {
	log := logging.S(r.Context())
	log.Info("GET /servers - ListServers called")

	// Parse query parameters
	region := r.URL.Query().Get("region")
	serverType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	// Set default limit if not provided or invalid
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Call repository to get servers
	servers, err := h.Service.ListServers(r.Context(), region, serverType, status, limit, offset)
	if err != nil {
		log.Errorw("Failed to list servers", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list servers")
		return
	}

	// Convert to response type
	var resp []*packets.ServerResponse
	for _, s := range servers {
		serverResp := &packets.ServerResponse{
			ID:     s.ID,
			Region: s.Region,
			Type:   s.Type,
			State:  s.State,
		}

		// Add IP address if available
		if s.IP != nil {
			serverResp.IPAddress = s.IP.Address
		}

		resp = append(resp, serverResp)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Errorw("Failed to encode response", "error", err)
	}
}

// @Summary Get server logs
// @Description Return last 100 lifecycle events (ring buffer)
// @Tags servers
// @Produce json
// @Param id path string true "Server ID"
// @Success 200 {array} eventLogResponse
// @Failure 404 {object} errorResponse
// @Router /servers/{id}/logs [get]
func (h *serverHandler) GetServerLogs(w http.ResponseWriter, r *http.Request) {
	log := logging.S(r.Context())
	id := chi.URLParam(r, "id")
	log.Infow("GET /servers/{id}/logs - GetServerLogs called", "id", id)
	events, err := h.Service.GetEvents(r.Context(), id, 100)
	if err != nil {
		log.Errorw("Failed to fetch logs for server", "id", id, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to fetch logs")
		return
	}
	if len(events) == 0 {
		log.Warnw("No logs found for server", "id", id)
		respondError(w, http.StatusNotFound, "no logs found")
		return
	}
	resp := make([]*packets.EventLogResponse, 0, len(events))
	for _, e := range events {
		resp = append(resp, &packets.EventLogResponse{Timestamp: e.Timestamp.Format("2006-01-02T15:04:05Z07:00"), Type: e.Type, Message: e.Message})
	}
	json.NewEncoder(w).Encode(resp)
}

type errorResponse struct {
	Error string `json:"error"`
}

func respondError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
