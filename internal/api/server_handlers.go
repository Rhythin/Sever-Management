package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rhythin/sever-management/internal/logging"
	"github.com/rhythin/sever-management/internal/persistence"
	"github.com/rhythin/sever-management/internal/service"
)

// ServerHandlers provides HTTP handlers for server endpoints

type ServerHandlers struct {
	Service *service.ServerService
	Repo    *persistence.ServerRepo
}

// @Summary Provision a new virtual server
// @Description Provision a new virtual server
// @Tags servers
// @Accept json
// @Produce json
// @Param server body provisionRequest true "Server spec"
// @Success 201 {object} provisionResponse
// @Failure 400 {object} errorResponse
// @Router /server [post]
func (h *ServerHandlers) ProvisionServer(w http.ResponseWriter, r *http.Request) {
	log := logging.S(r.Context())
	log.Infow("POST /server - ProvisionServer called")
	var req provisionRequest
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
	json.NewEncoder(w).Encode(provisionResponse{ID: id})
}

// @Summary Get server metadata
// @Description Retrieve full metadata for a server
// @Tags servers
// @Produce json
// @Param id path string true "Server ID"
// @Success 200 {object} serverResponse
// @Failure 404 {object} errorResponse
// @Router /servers/{id} [get]
func (h *ServerHandlers) GetServer(w http.ResponseWriter, r *http.Request) {
	log := logging.S(r.Context())
	id := chi.URLParam(r, "id")
	log.Infow("GET /servers/{id} - GetServer called", "id", id)
	ctx := r.Context()
	server, err := h.Repo.GetByID(ctx, id)
	if err != nil || server == nil {
		log.Warnw("Server not found", "id", id)
		respondError(w, http.StatusNotFound, "server not found")
		return
	}
	var lastBilled *string
	if server.Billing.LastBilledAt != nil {
		s := server.Billing.LastBilledAt.Format("2006-01-02T15:04:05Z07:00")
		lastBilled = &s
	}
	resp := serverResponse{
		ID:     server.ID,
		State:  server.State,
		Region: server.Region,
		Type:   server.Type,
		Billing: &billingResponse{
			AccumulatedSeconds: server.Billing.AccumulatedSeconds,
			LastBilledAt:       lastBilled,
			TotalCost:          server.Billing.TotalCost,
		},
	}
	json.NewEncoder(w).Encode(resp)
}

// @Summary Perform server action
// @Description Perform an action (start, stop, reboot, terminate) on a server
// @Tags servers
// @Accept json
// @Produce json
// @Param id path string true "Server ID"
// @Param action body actionRequest true "Action"
// @Success 200 {object} actionResponse
// @Failure 409 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /servers/{id}/action [post]
func (h *ServerHandlers) ServerAction(w http.ResponseWriter, r *http.Request) {
	log := logging.S(r.Context())
	id := chi.URLParam(r, "id")
	log.Infow("POST /servers/{id}/action - ServerAction called", "id", id)
	var req actionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Warnw("Invalid action request body", "error", err)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	err := h.Service.Action(r.Context(), id, req.Action)
	if err != nil {
		if err.Error() == "server not found" {
			log.Warnw("Server not found for action", "id", id)
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Warnw("Invalid FSM transition for server", "id", id, "error", err)
		respondError(w, http.StatusConflict, err.Error())
		return
	}
	log.Infow("Action performed on server", "action", req.Action, "id", id)
	json.NewEncoder(w).Encode(actionResponse{Result: "ok"})
}

// @Summary List servers
// @Description List all servers, filterable by region, status, type; supports pagination and sorting
// @Tags servers
// @Produce json
// @Param region query string false "Region"
// @Param status query string false "Status"
// @Param type query string false "Type"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} serverResponse
// @Router /servers [get]
func (h *ServerHandlers) ListServers(w http.ResponseWriter, r *http.Request) {
	log := logging.S(r.Context())
	log.Infow("GET /servers - ListServers called")
	q := r.URL.Query()
	region := q.Get("region")
	status := q.Get("status")
	typ := q.Get("type")
	limit := 20
	offset := 0
	if l := q.Get("limit"); l != "" {
		fmt.Sscanf(l, "%d", &limit)
	}
	if o := q.Get("offset"); o != "" {
		fmt.Sscanf(o, "%d", &offset)
	}
	servers, err := h.Repo.List(r.Context(), region, status, typ, limit, offset)
	if err != nil {
		log.Errorw("Failed to list servers", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list servers")
		return
	}
	resp := make([]serverResponse, 0, len(servers))
	for _, s := range servers {
		resp = append(resp, serverResponse{ID: s.ID, State: s.State})
	}
	json.NewEncoder(w).Encode(resp)
}

// @Summary Get server logs
// @Description Return last 100 lifecycle events (ring buffer)
// @Tags servers
// @Produce json
// @Param id path string true "Server ID"
// @Success 200 {array} eventLogResponse
// @Failure 404 {object} errorResponse
// @Router /servers/{id}/logs [get]
func (h *ServerHandlers) GetServerLogs(w http.ResponseWriter, r *http.Request) {
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
	resp := make([]eventLogResponse, 0, len(events))
	for _, e := range events {
		resp = append(resp, eventLogResponse{Timestamp: e.Timestamp.Format("2006-01-02T15:04:05Z07:00"), Type: e.Type, Message: e.Message})
	}
	json.NewEncoder(w).Encode(resp)
}

// --- Request/Response types ---
type provisionRequest struct {
	Region string `json:"region"`
	Type   string `json:"type"`
}
type provisionResponse struct {
	ID string `json:"id"`
}
type serverResponse struct {
	ID      string           `json:"id"`
	State   string           `json:"state"`
	Region  string           `json:"region,omitempty"`
	Type    string           `json:"type,omitempty"`
	Billing *billingResponse `json:"billing,omitempty"`
}

type billingResponse struct {
	AccumulatedSeconds int64   `json:"accumulated_seconds"`
	LastBilledAt       *string `json:"last_billed_at,omitempty"`
	TotalCost          float64 `json:"total_cost"`
}
type errorResponse struct {
	Error string `json:"error"`
}
type actionRequest struct {
	Action string `json:"action"` // must be one of start|stop|reboot|terminate
}
type actionResponse struct {
	Result string `json:"result"`
}
type eventLogResponse struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Message   string `json:"message"`
}

func respondError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorResponse{Error: msg})
}
