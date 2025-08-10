package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rhythin/sever-management/internal/domain"
	"github.com/rhythin/sever-management/internal/packets"
	"github.com/rhythin/sever-management/internal/persistence"
	"github.com/stretchr/testify/assert"
)

// Mock implementations for testing
type mockServerService struct {
	shouldFailProvision bool
	shouldFailAction    bool
	shouldFailGetEvents bool
	provisionedID       string
}

func (m *mockServerService) Provision(ctx context.Context, region, typ string) (string, error) {
	if m.shouldFailProvision {
		return "", assert.AnError
	}
	return m.provisionedID, nil
}

func (m *mockServerService) Action(ctx context.Context, id, action string) error {
	if m.shouldFailAction {
		return assert.AnError
	}
	return nil
}

func (m *mockServerService) GetEvents(ctx context.Context, id string, n int) ([]domain.EventLogEntry, error) {
	if m.shouldFailGetEvents {
		return nil, assert.AnError
	}
	return []domain.EventLogEntry{
		{
			Timestamp: time.Now(),
			Type:      domain.EventStarted,
			Message:   "test message",
		},
	}, nil
}

type mockServerRepo struct {
	shouldFailGetByID bool
	shouldFailList    bool
	server            *persistence.Server
	servers           []*persistence.Server
}

func (m *mockServerRepo) GetByID(ctx context.Context, id string) (*persistence.Server, error) {
	if m.shouldFailGetByID {
		return nil, assert.AnError
	}
	return m.server, nil
}

func (m *mockServerRepo) List(ctx context.Context, region, status, typ string, limit, offset int) ([]*persistence.Server, error) {
	if m.shouldFailList {
		return nil, assert.AnError
	}
	return m.servers, nil
}

func (m *mockServerRepo) Create(ctx context.Context, server *persistence.Server) error {
	return nil
}

func (m *mockServerRepo) UpdateState(ctx context.Context, id string, state string) error {
	return nil
}

func (m *mockServerRepo) UpdateTimestamps(ctx context.Context, id string, started, stopped, terminated *time.Time) error {
	return nil
}

func (m *mockServerRepo) UpdateServer(ctx context.Context, id string, updates *persistence.Server) error {
	return nil
}

func (m *mockServerRepo) UpdateBilling(ctx context.Context, id string, accumulatedSeconds int64, totalCost float64) error {
	return nil
}

func TestServerHandlers_ProvisionServer_Success(t *testing.T) {
	service := &mockServerService{provisionedID: "server-123"}
	repo := &mockServerRepo{}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	req := packets.ProvisionRequest{
		Region: "us-west-1",
		Type:   "t2.micro",
	}
	body, _ := json.Marshal(req)

	request := httptest.NewRequest("POST", "/server", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handlers.ProvisionServer(response, request)

	assert.Equal(t, http.StatusCreated, response.Code)

	var resp packets.ProvisionResponse
	err := json.NewDecoder(response.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "server-123", resp.ID)
}

func TestServerHandlers_ProvisionServer_InvalidRequest(t *testing.T) {
	service := &mockServerService{}
	repo := &mockServerRepo{}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	// Test missing region
	req := packets.ProvisionRequest{
		Type: "t2.micro",
	}
	body, _ := json.Marshal(req)

	request := httptest.NewRequest("POST", "/server", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handlers.ProvisionServer(response, request)

	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestServerHandlers_ProvisionServer_ServiceFailure(t *testing.T) {
	service := &mockServerService{shouldFailProvision: true}
	repo := &mockServerRepo{}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	req := packets.ProvisionRequest{
		Region: "us-west-1",
		Type:   "t2.micro",
	}
	body, _ := json.Marshal(req)

	request := httptest.NewRequest("POST", "/server", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	handlers.ProvisionServer(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestServerHandlers_GetServer_Success(t *testing.T) {
	now := time.Now()
	server := &persistence.Server{
		ID:     "server-123",
		State:  "running",
		Region: "us-west-1",
		Type:   "t2.micro",
		Billing: &persistence.Billing{
			AccumulatedSeconds: 3600,
			LastBilledAt:       &now,
			TotalCost:          0.10,
		},
	}

	service := &mockServerService{}
	repo := &mockServerRepo{server: server}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	request := httptest.NewRequest("GET", "/servers/server-123", nil)
	response := httptest.NewRecorder()

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "server-123")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	handlers.GetServer(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp packets.ServerResponse
	err := json.NewDecoder(response.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "server-123", resp.ID)
	assert.Equal(t, "running", resp.State)
	assert.Equal(t, "us-west-1", resp.Region)
	assert.Equal(t, "t2.micro", resp.Type)
	assert.Equal(t, int64(3600), resp.Billing.AccumulatedSeconds)
	assert.Equal(t, 0.10, resp.Billing.TotalCost)
}

func TestServerHandlers_GetServer_NotFound(t *testing.T) {
	service := &mockServerService{}
	repo := &mockServerRepo{shouldFailGetByID: true}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	request := httptest.NewRequest("GET", "/servers/non-existent", nil)
	response := httptest.NewRecorder()

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "non-existent")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	handlers.GetServer(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestServerHandlers_ServerAction_Success(t *testing.T) {
	service := &mockServerService{}
	repo := &mockServerRepo{}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	req := packets.ActionRequest{Action: "start"}
	body, _ := json.Marshal(req)

	request := httptest.NewRequest("POST", "/servers/server-123/action", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "server-123")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	handlers.ServerAction(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp packets.ActionResponse
	err := json.NewDecoder(response.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "ok", resp.Result)
}

func TestServerHandlers_ServerAction_ServiceFailure(t *testing.T) {
	service := &mockServerService{shouldFailAction: true}
	repo := &mockServerRepo{}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	req := packets.ActionRequest{Action: "start"}
	body, _ := json.Marshal(req)

	request := httptest.NewRequest("POST", "/servers/server-123/action", bytes.NewBuffer(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "server-123")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	handlers.ServerAction(response, request)

	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestServerHandlers_ListServers_Success(t *testing.T) {
	servers := []*persistence.Server{
		{ID: "server-1", State: "running"},
		{ID: "server-2", State: "stopped"},
	}

	service := &mockServerService{}
	repo := &mockServerRepo{servers: servers}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	request := httptest.NewRequest("GET", "/servers", nil)
	response := httptest.NewRecorder()

	handlers.ListServers(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp []packets.ServerResponse
	err := json.NewDecoder(response.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "server-1", resp[0].ID)
	assert.Equal(t, "running", resp[0].State)
}

func TestServerHandlers_ListServers_WithQueryParams(t *testing.T) {
	servers := []*persistence.Server{
		{ID: "server-1", State: "running"},
	}

	service := &mockServerService{}
	repo := &mockServerRepo{servers: servers}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	request := httptest.NewRequest("GET", "/servers?region=us-west-1&status=running&limit=10&offset=0", nil)
	response := httptest.NewRecorder()

	handlers.ListServers(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
}

func TestServerHandlers_ListServers_RepoFailure(t *testing.T) {
	service := &mockServerService{}
	repo := &mockServerRepo{shouldFailList: true}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	request := httptest.NewRequest("GET", "/servers", nil)
	response := httptest.NewRecorder()

	handlers.ListServers(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestServerHandlers_GetServerLogs_Success(t *testing.T) {
	service := &mockServerService{}
	repo := &mockServerRepo{}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	request := httptest.NewRequest("GET", "/servers/server-123/logs", nil)
	response := httptest.NewRecorder()

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "server-123")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	handlers.GetServerLogs(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var resp []packets.EventLogResponse
	err := json.NewDecoder(response.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)
	assert.Equal(t, "test", resp[0].Type)
	assert.Equal(t, "test message", resp[0].Message)
}

func TestServerHandlers_GetServerLogs_ServiceFailure(t *testing.T) {
	service := &mockServerService{shouldFailGetEvents: true}
	repo := &mockServerRepo{}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	request := httptest.NewRequest("GET", "/servers/server-123/logs", nil)
	response := httptest.NewRecorder()

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "server-123")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	handlers.GetServerLogs(response, request)

	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestServerHandlers_GetServerLogs_NoLogs(t *testing.T) {
	service := &mockServerService{shouldFailGetEvents: false}
	repo := &mockServerRepo{}
	handlers := &ServerHandlers{Service: service, Repo: repo}

	// Override the GetEvents method for this test
	originalGetEvents := service.GetEvents
	service.GetEvents = func(ctx context.Context, id string, n int) ([]packets.EventLog, error) {
		return []packets.EventLog{}, nil
	}
	defer func() { service.GetEvents = originalGetEvents }()

	request := httptest.NewRequest("GET", "/servers/server-123/logs", nil)
	response := httptest.NewRecorder()

	// Set up chi context with URL parameters
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "server-123")
	request = request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, rctx))

	handlers.GetServerLogs(response, request)

	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestRespondError(t *testing.T) {
	response := httptest.NewRecorder()

	respondError(response, http.StatusBadRequest, "test error")

	assert.Equal(t, http.StatusBadRequest, response.Code)

	var resp errorResponse
	err := json.NewDecoder(response.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "test error", resp.Error)
}
