package domain

import (
	"context"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/rhythin/sever-management/internal/logging"
)

// ServerState represents the finite states a server can be in
// Use string for easy JSON and DB mapping

type ServerState string

const (
	ServerProvisioning ServerState = "provisioning"
	ServerRunning      ServerState = "running"
	ServerStopped      ServerState = "stopped"
	ServerRebooting    ServerState = "rebooting"
	ServerTerminated   ServerState = "terminated"
)

// ServerType for billing and resource simulation

type ServerType string

const (
	TypeT2Micro ServerType = "t2.micro"
	TypeT2Small ServerType = "t2.small"
)

// Server represents a virtual server instance

type Server struct {
	ID           string // UUID
	Region       string
	Type         ServerType
	IP           net.IP
	State        ServerState
	CreatedAt    time.Time
	UpdatedAt    time.Time
	StartedAt    *time.Time // for uptime/billing
	StoppedAt    *time.Time // for idle reaper
	TerminatedAt *time.Time
	Billing      BillingInfo
	Log          *EventRingBuffer
	mu           sync.Mutex // protects FSM transitions, timestamps, billing
}

// BillingInfo tracks cost and uptime

type BillingInfo struct {
	AccumulatedSeconds int64 // total uptime in seconds
	LastBilledAt       *time.Time
	TotalCost          float64
}

// EventType for server lifecycle events

type EventType string

const (
	EventProvisioned EventType = "provisioned"
	EventStarted     EventType = "started"
	EventStopped     EventType = "stopped"
	EventRebooted    EventType = "rebooted"
	EventTerminated  EventType = "terminated"
	EventBilled      EventType = "billed"
	EventReaped      EventType = "reaped"
)

// EventLogEntry represents a single server event

type EventLogEntry struct {
	Timestamp time.Time
	Type      EventType
	Message   string
}

// EventRingBuffer holds the last N events (thread-safe)

type EventRingBuffer struct {
	buf   []EventLogEntry
	size  int
	next  int
	count int
	mu    sync.Mutex
}

func NewEventRingBuffer(size int) *EventRingBuffer {
	return &EventRingBuffer{buf: make([]EventLogEntry, size), size: size}
}

func (r *EventRingBuffer) Add(e EventLogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.next] = e
	r.next = (r.next + 1) % r.size
	if r.count < r.size {
		r.count++
	}
}

func (r *EventRingBuffer) List() []EventLogEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	res := make([]EventLogEntry, 0, r.count)
	for i := 0; i < r.count; i++ {
		idx := (r.next - r.count + i + r.size) % r.size
		res = append(res, r.buf[idx])
	}
	return res
}

// FSM transition logic (thread-safe)

var ErrInvalidTransition = errors.New("invalid state transition")

func (s *Server) Transition(ctx context.Context, action string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	log := logging.S(ctx)
	now := time.Now()
	log.Infow("FSM transition attempt", "server_id", s.ID, "from", s.State, "action", action)
	switch action {
	case "start":
		if s.State == ServerStopped {
			s.State = ServerRunning
			s.StartedAt = &now
			s.Log.Add(EventLogEntry{Timestamp: now, Type: EventStarted, Message: "Server started"})
			log.Infow("FSM transition success", "server_id", s.ID, "to", s.State)
			return nil
		}
	case "stop":
		if s.State == ServerRunning {
			s.State = ServerStopped
			s.StoppedAt = &now
			s.Log.Add(EventLogEntry{Timestamp: now, Type: EventStopped, Message: "Server stopped"})
			log.Infow("FSM transition success", "server_id", s.ID, "to", s.State)
			return nil
		}
	case "reboot":
		if s.State == ServerRunning {
			s.State = ServerRebooting
			s.Log.Add(EventLogEntry{Timestamp: now, Type: EventRebooted, Message: "Server rebooting"})
			// Simulate reboot: immediately back to running
			s.State = ServerRunning
			s.Log.Add(EventLogEntry{Timestamp: now, Type: EventStarted, Message: "Server rebooted and running"})
			log.Infow("FSM transition success", "server_id", s.ID, "to", s.State)
			return nil
		}
	case "terminate":
		if s.State != ServerTerminated {
			s.State = ServerTerminated
			s.TerminatedAt = &now
			s.Log.Add(EventLogEntry{Timestamp: now, Type: EventTerminated, Message: "Server terminated"})
			log.Infow("FSM transition success", "server_id", s.ID, "to", s.State)
			return nil
		}
	}
	log.Warnw("FSM invalid transition", "server_id", s.ID, "from", s.State, "action", action)
	return ErrInvalidTransition
}
