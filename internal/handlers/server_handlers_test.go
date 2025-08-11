package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rhythin/sever-management/internal/domain"
	"github.com/rhythin/sever-management/internal/packets"
	"github.com/rhythin/sever-management/internal/persistence"
	"github.com/rhythin/sever-management/internal/service"
	"github.com/stretchr/testify/mock"

	mockService "github.com/rhythin/sever-management/internal/service/mocks"
)

func GenerateProvisionServerRequest(region string, body interface{}) *http.Request {

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil
	}

	req, err := http.NewRequest("POST", "/server", strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil
	}

	return req
}

func Test_serverHandler_ProvisionServer(t *testing.T) {
	mockService := &mockService.ServerService{}
	mockService.On("Provision", mock.Anything, "1", "type").Return("1", nil)
	mockService.On("Provision", mock.Anything, "2", "type").Return("1", errors.New("service layer error"))
	mockService.On("Provision", mock.Anything, "3", "type").Return("", errors.New("no available IPs"))
	type fields struct {
		Service service.ServerService
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "ProvisionServer success",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateProvisionServerRequest("region", packets.ProvisionRequest{Region: "1", Type: "type"}),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "ProvisionServer error service Layer",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateProvisionServerRequest("region", packets.ProvisionRequest{Region: "2", Type: "type"}),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "ProvisionServer invalid request body missing region and type",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateProvisionServerRequest("", packets.ProvisionRequest{Region: "", Type: ""}),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "ProvisionServer invalid request body",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateProvisionServerRequest("", map[string]interface{}{"region": 32, "type": 324}),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "ProvisionServer no available IPs",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateProvisionServerRequest("region", packets.ProvisionRequest{Region: "3", Type: "type"}),
			},
			fields: fields{
				Service: mockService,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &serverHandler{
				Service: tt.fields.Service,
			}
			h.ProvisionServer(tt.args.w, tt.args.r)
		})
	}
}

func GenerateGetServerRequest(id string) *http.Request {

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)

	req, err := http.NewRequest("GET", "/servers/"+id, nil)
	if err != nil {
		return nil
	}

	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	return req
}

func Test_serverHandler_GetServer(t *testing.T) {

	mockService := &mockService.ServerService{}
	mockService.On("GetServerByID", mock.Anything, "1").Return(&persistence.Server{
		ID:     "1",
		Region: "region",
		Type:   "type",
		State:  "state",
		IP:     &persistence.IPAddress{Address: "127.0.0.1"},
	}, nil)
	mockService.On("GetServerByID", mock.Anything, "2").Return(nil, errors.New("service layer error"))
	type fields struct {
		Service service.ServerService
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "GetServer success",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateGetServerRequest("1"),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "GetServer error service Layer",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateGetServerRequest("2"),
			},
			fields: fields{
				Service: mockService,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &serverHandler{
				Service: tt.fields.Service,
			}
			h.GetServer(tt.args.w, tt.args.r)
		})
	}
}

func GenerateServerActionRequestGenerator(id string, body interface{}) *http.Request {

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil
	}

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)

	req, err := http.NewRequest("POST", "/servers/"+id+"/action", strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil
	}

	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	return req
}

func Test_serverHandler_ServerAction(t *testing.T) {

	mockService := &mockService.ServerService{}
	mockService.On("Action", mock.Anything, "1", domain.ServerAction("start")).Return(nil)
	mockService.On("Action", mock.Anything, "2", domain.ServerAction("start")).Return(errors.New("service layer error"))
	mockService.On("Action", mock.Anything, "3", domain.ServerAction("start")).Return(errors.New("invalid state transition"))
	mockService.On("Action", mock.Anything, "4", domain.ServerAction("start")).Return(nil)
	type fields struct {
		Service service.ServerService
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "ServerAction success",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateServerActionRequestGenerator("1", packets.ActionRequest{Action: "start"}),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "ServerAction invalid action",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateServerActionRequestGenerator("1", packets.ActionRequest{Action: "invalid"}),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "ServerAction error service Layer",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateServerActionRequestGenerator("2", packets.ActionRequest{Action: "start"}),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "ServerAction invalid service layer error",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateServerActionRequestGenerator("3", packets.ActionRequest{Action: "start"}),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "invalid request body",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateServerActionRequestGenerator("invalid", map[string]interface{}{"action": 1}),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "missing id",
			args: args{
				w: httptest.NewRecorder(),
				r: GenerateServerActionRequestGenerator("", packets.ActionRequest{Action: "start"}),
			},
			fields: fields{
				Service: mockService,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &serverHandler{
				Service: tt.fields.Service,
			}
			h.ServerAction(tt.args.w, tt.args.r)
		})
	}
}

func GetServerRequestGenerator(region string, serverType string, status string, limit int, offset int) *http.Request {

	return httptest.NewRequest("GET", "/servers?region="+region+"&type="+serverType+"&status="+status+"&limit="+strconv.Itoa(limit)+"&offset="+strconv.Itoa(offset), nil)
}

func Test_serverHandler_ListServers(t *testing.T) {

	svcResp := []*persistence.Server{
		{
			ID: "1",
			IP: &persistence.IPAddress{Address: "127.0.0.1"},
		},
		{
			ID: "2",
			IP: &persistence.IPAddress{Address: "127.0.0.2"},
		},
	}

	mockService := &mockService.ServerService{}
	mockService.On("ListServers", mock.Anything, "1", "type", "status", 20, 0).Return(svcResp, nil)
	mockService.On("ListServers", mock.Anything, "2", "type", "status", 20, 0).Return(nil, errors.New("service layer error"))
	mockService.On("ListServers", mock.Anything, "3", "type", "status", 20, 0).Return(svcResp, nil)

	type fields struct {
		Service service.ServerService
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "ListServers success",
			args: args{
				w: httptest.NewRecorder(),
				r: GetServerRequestGenerator("1", "type", "status", 0, 0),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "ListServers error service Layer",
			args: args{
				w: httptest.NewRecorder(),
				r: GetServerRequestGenerator("2", "type", "status", 20, 0),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "ListServers no servers found",
			args: args{
				w: httptest.NewRecorder(),
				r: GetServerRequestGenerator("3", "type", "status", 20, 0),
			},
			fields: fields{
				Service: mockService,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &serverHandler{
				Service: tt.fields.Service,
			}
			h.ListServers(tt.args.w, tt.args.r)
		})
	}
}

func GetLogsRequestGenerator(id string) *http.Request {

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)

	req, err := http.NewRequest("GET", "/servers/"+id+"/logs", nil)
	if err != nil {
		return nil
	}

	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
	req = req.WithContext(ctx)

	return req
}

func Test_serverHandler_GetServerLogs(t *testing.T) {

	mockService := &mockService.ServerService{}
	mockService.On("GetEvents", mock.Anything, "1", 100).Return([]persistence.EventLog{{}, {}}, nil)
	mockService.On("GetEvents", mock.Anything, "2", 100).Return(nil, errors.New("service layer error"))
	mockService.On("GetEvents", mock.Anything, "3", 100).Return([]persistence.EventLog{}, nil)

	type fields struct {
		Service service.ServerService
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "GetServerLogs success",
			args: args{
				w: httptest.NewRecorder(),
				r: GetLogsRequestGenerator("1"),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "GetServerLogs error service Layer",
			args: args{
				w: httptest.NewRecorder(),
				r: GetLogsRequestGenerator("2"),
			},
			fields: fields{
				Service: mockService,
			},
		},
		{
			name: "GetServerLogs no logs found",
			args: args{
				w: httptest.NewRecorder(),
				r: GetLogsRequestGenerator("3"),
			},
			fields: fields{
				Service: mockService,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &serverHandler{
				Service: tt.fields.Service,
			}
			h.GetServerLogs(tt.args.w, tt.args.r)
		})
	}
}

func Test_respondError(t *testing.T) {
	type args struct {
		w    http.ResponseWriter
		code int
		msg  string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
		{
			name: "respondError",
			args: args{
				w:    httptest.NewRecorder(),
				code: http.StatusInternalServerError,
				msg:  "Internal Server Error",
			},
		},
		{
			name: "respondError with message",
			args: args{
				w:    httptest.NewRecorder(),
				code: http.StatusUnprocessableEntity,
				msg:  "Invalid request",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			respondError(tt.args.w, tt.args.code, tt.args.msg)
		})
	}
}
