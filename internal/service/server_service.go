package service

import (
	"context"
	"errors"
	"time"

	"github.com/rhythin/sever-management/internal/domain"
	"github.com/rhythin/sever-management/internal/persistence"
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
	server, err := s.servers.GetByID(ctx, id)
	if err != nil || server == nil {
		return errors.New("server not found")
	}
	d := toDomainServer(server)
	if err := d.Transition(action); err != nil {
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
		return err
	}
	if err := s.servers.UpdateTimestamps(ctx, id, started, stopped, terminated); err != nil {
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
