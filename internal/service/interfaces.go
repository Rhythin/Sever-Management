package service

import (
	"context"

	"github.com/rhythin/sever-management/internal/domain"
	"github.com/rhythin/sever-management/internal/persistence"
)

// ServerService defines the interface for server operations needed by handlers
type ServerService interface {
	Provision(ctx context.Context, region, typ string) (string, error)
	Action(ctx context.Context, id string, action domain.ServerAction) error
	GetEvents(ctx context.Context, id string, n int) ([]persistence.EventLog, error)
	ListServers(ctx context.Context, region, status, typ string, limit, offset int) ([]*persistence.Server, error)
	GetServerByID(ctx context.Context, id string) (*persistence.Server, error)
}
