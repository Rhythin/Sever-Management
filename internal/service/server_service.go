package service

import (
	"context"
	"errors"
	"time"

	"github.com/rhythin/sever-management/internal/domain"
	"github.com/rhythin/sever-management/internal/logging"
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
func (s *ServerService) Action(ctx context.Context, id string, action string) error {
	log := logging.S(ctx)
	log.Infow("ServerService.Action called", "id", id, "action", action)
	server, err := s.servers.GetByID(ctx, id)
	if err != nil || server == nil {
		log.Warnw("Server not found", "id", id)
		return errors.New("server not found")
	}
	d := toDomainServer(server)
	if err := d.Transition(ctx, domain.ServerAction(action)); err != nil {
		log.Warnw("Invalid FSM transition for server", "id", id, "error", err)
		return err
	}
	// Persist state and timestamps
	var started, stopped, terminated *time.Time
	switch domain.ServerAction(action) {
	case domain.ActionStart:
		started = d.StartedAt
	case domain.ActionStop:
		stopped = d.StoppedAt
	case domain.ActionTerminate:
		terminated = d.TerminatedAt
	}
	if err := s.servers.UpdateState(ctx, id, string(d.State)); err != nil {
		log.Errorw("Failed to update state for server", "id", id, "error", err)
		return err
	}
	if err := s.servers.UpdateTimestamps(ctx, id, started, stopped, terminated); err != nil {
		log.Errorw("Failed to update timestamps for server", "id", id, "error", err)
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
	log.Infow("Action performed on server", "action", action, "id", id)
	return nil
}

// Provision provisions a new server, allocates an IP, and persists it
func (s *ServerService) Provision(ctx context.Context, region, typ string) (string, error) {
	log := logging.S(ctx)
	log.Infow("ServerService.Provision called", "region", region, "type", typ)

	// Allocate IP
	ip, err := s.ips.AllocateIP(ctx)
	if err != nil {
		log.Errorw("Failed to allocate IP", "error", err)
		return "", err
	}
	if ip == nil {
		log.Warnw("No available IPs for provisioning")
		return "", errors.New("no available IPs")
	}

	// Create server model
	now := time.Now()
	server := &persistence.Server{
		Region:    region,
		Type:      typ,
		IPID:      &ip.ID,
		State:     string(domain.ServerProvisioning),
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Persist server
	err = s.servers.Create(ctx, server)
	if err != nil {
		log.Errorw("Failed to persist server", "error", err)
		// Rollback IP allocation
		_ = s.ips.ReleaseIP(ctx, ip.ID)
		return "", err
	}

	// Link IP to server record for reverse reference
	_ = s.ips.AssignIPToServer(ctx, ip.ID, server.ID)

	// Optionally, update server state to running immediately (simulate provisioning)
	err = s.servers.UpdateState(ctx, server.ID, string(domain.ServerRunning))
	if err != nil {
		log.Errorw("Failed to update server state to running", "error", err)
		return server.ID, err // server is created, but state update failed
	}

	// Log provision and running events
	_ = s.events.Append(ctx, &persistence.EventLog{
		ServerID:  server.ID,
		Timestamp: now,
		Type:      string(domain.EventProvisioned),
		Message:   "Server provisioned",
	})
	_ = s.events.Append(ctx, &persistence.EventLog{
		ServerID:  server.ID,
		Timestamp: now,
		Type:      string(domain.EventStarted),
		Message:   "Server running",
	})

	return server.ID, nil
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
