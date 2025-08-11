package persistence

import (
	"context"
	"time"
)

// ServerRepoInterface defines the interface for server repository operations
type ServerRepoInterface interface {
	GetByID(ctx context.Context, id string) (*Server, error)
	Create(ctx context.Context, server *Server) error
	UpdateState(ctx context.Context, id string, state string) error
	UpdateTimestamps(ctx context.Context, id string, started, stopped, terminated *time.Time) error
	UpdateServer(ctx context.Context, id string, updates *Server) error
	UpdateBilling(ctx context.Context, id string, accumulatedSeconds int64, totalCost float64) error
	List(ctx context.Context, region, status, typ string, limit, offset int) ([]*Server, error)
}

// IPRepoInterface defines the interface for IP repository operations
type IPRepoInterface interface {
	AllocateIP(ctx context.Context) (*IPAddress, error)
	ReleaseIP(ctx context.Context, id uint) error
	AssignIPToServer(ctx context.Context, ipID uint, serverID string) error
}

// EventRepoInterface defines the interface for event repository operations
type EventRepoInterface interface {
	Append(ctx context.Context, event *EventLog) error
	LastN(ctx context.Context, serverID string, n int) ([]EventLog, error)
}
