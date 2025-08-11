package persistence

import (
	"context"

	"github.com/rhythin/sever-management/internal/logging"
	"gorm.io/gorm"
)

// EventRepo handles event log persistence

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) EventRepoInterface {
	return &EventRepo{db: db}
}

func (r *EventRepo) Append(ctx context.Context, event *EventLog) error {
	log := logging.S(ctx)
	log.Debugw("EventRepo.Append called", "serverID", event.ServerID, "type", event.Type)
	err := r.db.WithContext(ctx).Create(event).Error
	if err != nil {
		log.Errorw("EventRepo.Append failed", "serverID", event.ServerID, "error", err)
	}
	return err
}

func (r *EventRepo) LastN(ctx context.Context, serverID string, n int) ([]EventLog, error) {
	log := logging.S(ctx)
	log.Debugw("EventRepo.LastN called", "serverID", serverID, "n", n)
	var events []EventLog
	err := r.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Order("timestamp DESC").
		Limit(n).
		Find(&events).Error
	if err != nil {
		log.Errorw("EventRepo.LastN failed", "serverID", serverID, "error", err)
	}
	return events, err
}

// GetEvents returns all events for a server, ordered by timestamp (newest first)
func (r *EventRepo) GetEvents(ctx context.Context, serverID string) ([]EventLog, error) {
	log := logging.S(ctx)
	log.Debugw("EventRepo.GetEvents called", "serverID", serverID)
	var events []EventLog
	err := r.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Order("timestamp DESC").
		Find(&events).Error
	if err != nil {
		log.Errorw("EventRepo.GetEvents failed", "serverID", serverID, "error", err)
	}
	return events, err
}

// AddEvent is an alias for Append for backward compatibility
func (r *EventRepo) AddEvent(ctx context.Context, event *EventLog) error {
	return r.Append(ctx, event)
}
