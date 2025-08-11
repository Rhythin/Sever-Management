package handlers

import (
	"net/http"

	"github.com/rhythin/sever-management/internal/service"
)

type ServerHandler interface {
	ProvisionServer(w http.ResponseWriter, r *http.Request)
	ServerAction(w http.ResponseWriter, r *http.Request)
	GetServerLogs(w http.ResponseWriter, r *http.Request)
	GetServer(w http.ResponseWriter, r *http.Request)
	ListServers(w http.ResponseWriter, r *http.Request)
}

func NewServerHandler(service service.ServerService) ServerHandler {
	return &serverHandler{Service: service}
}
