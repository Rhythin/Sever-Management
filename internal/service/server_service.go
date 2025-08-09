package service

import (
	"context"
	"errors"
	"time"

	"github.com/rhythin/sever-management/internal/domain"
	"github.com/rhythin/sever-management/internal/persistence"
	"go.uber.org/zap"
)

// ServerService orchestrates server FSM and actions

type ServerService struct {
	servers *persistence.ServerRepo
	ips     *persistence.IPRepo
	events  *persistence.EventRepo
}

func NewServerService(servers *persistence.ServerRepo, ips *persistence.IPRepo, events *persistence.EventRepo) *ServerService {
	return &ServerService{servers: servers, ips: ips, events: events}
}

// Action performs a state transition (start, stop, reboot, terminate)
func (s *ServerService) Action(ctx context.Context, id, action string) error {
	zap.S().Infow("ServerService.Action called", "id", id, "action", action)
	server, err := s.servers.GetByID(ctx, id)
	if err != nil || server == nil {
		zap.S().Warnw("Server not found", "id", id)
		return errors.New("server not found")
	}
	d := toDomainServer(server)
	if err := d.Transition(action); err != nil {
		zap.S().Warnw("Invalid FSM transition for server", "id", id, "error", err)
		return err
	}
	// Persist state and timestamps
	var started, stopped, terminated *time.Time
	switch action {
	case "start":
		started = d.StartedAt
	case "stop":
		stopped = d.StoppedAt
	case "terminate":
		terminated = d.TerminatedAt
	}
	if err := s.servers.UpdateState(ctx, id, string(d.State)); err != nil {
		zap.S().Errorw("Failed to update state for server", "id", id, "error", err)
		return err
	}
	if err := s.servers.UpdateTimestamps(ctx, id, started, stopped, terminated); err != nil {
		zap.S().Errorw("Failed to update timestamps for server", "id", id, "error", err)
		return err
	}
	// Log event
	for _, e := range d.Log.List() {
		_ = s.events.Append(ctx, &persistence.EventLog{
			ServerID:  id,
			Timestamp: e.Timestamp,
			Type:      string(e.Type),
			Message:   e.Message,
		})
	}
	zap.S().Infow("Action performed on server", "action", action, "id", id)
	return nil
}

// toDomainServer maps persistence.Server to domain.Server (minimal for FSM)
func toDomainServer(s *persistence.Server) *domain.Server {
	return &domain.Server{
		ID:           s.ID,
		State:        domain.ServerState(s.State),
		StartedAt:    s.StartedAt,
		StoppedAt:    s.StoppedAt,
		TerminatedAt: s.TerminatedAt,
		Log:          domain.NewEventRingBuffer(100), // Not persisted, just for FSM
	}
}

func (s *ServerService) GetEvents(ctx context.Context, id string, n int) ([]persistence.EventLog, error) {
	return s.events.LastN(ctx, id, n)
}
