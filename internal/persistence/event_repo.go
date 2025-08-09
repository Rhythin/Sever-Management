package persistence

import (
	"context"

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
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *EventRepo) LastN(ctx context.Context, serverID string, n int) ([]EventLog, error) {
	var events []EventLog
	err := r.db.WithContext(ctx).
		Where("server_id = ?", serverID).
		Order("timestamp DESC").
		Limit(n).
		Find(&events).Error
	return events, err
}
