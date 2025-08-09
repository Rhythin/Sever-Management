package persistence

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// EventRepo handles event log persistence

type EventRepo struct {
	db *gorm.DB
}

func NewEventRepo(db *gorm.DB) *EventRepo {
	return &EventRepo{db: db}
}

func (r *EventRepo) Append(ctx context.Context, event *EventLog) error {
	zap.S().Debugw("EventRepo.Append called", "serverID", event.ServerID, "type", event.Type)
	err := r.db.WithContext(ctx).Create(event).Error
	if err != nil {
		zap.S().Errorw("EventRepo.Append failed", "serverID", event.ServerID, "error", err)
	}
	return err
}

func (r *EventRepo) LastN(ctx context.Context, serverID string, n int) ([]EventLog, error) {
	zap.S().Debugw("EventRepo.LastN called", "serverID", serverID, "n", n)
	var events []EventLog
	err := r.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Order("timestamp DESC").
		Limit(n).
		Find(&events).Error
	if err != nil {
		zap.S().Errorw("EventRepo.LastN failed", "serverID", serverID, "error", err)
	}
	return events, err
}
