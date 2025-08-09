package persistence

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ServerRepo handles server persistence

type ServerRepo struct {
	db *gorm.DB
}

func NewServerRepo(db *gorm.DB) *ServerRepo {
	return &ServerRepo{db: db}
}

func (r *ServerRepo) Create(ctx context.Context, s *Server) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *ServerRepo) GetByID(ctx context.Context, id string) (*Server, error) {
	var s Server
	err := r.db.WithContext(ctx).Preload("IP").Preload("Billing").Preload("Events").First(&s, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *ServerRepo) UpdateState(ctx context.Context, id string, newState string) error {
	return r.db.WithContext(ctx).Model(&Server{}).Where("id = ?", id).Update("state", newState).Error
}

func (r *ServerRepo) List(ctx context.Context, region, status, typ string, limit, offset int) ([]Server, error) {
	var servers []Server
	q := r.db.WithContext(ctx).Model(&Server{}).Preload("IP").Preload("Billing").Preload("Events")
	if region != "" {
		q = q.Where("region = ?", region)
	}
	if status != "" {
		q = q.Where("state = ?", status)
	}
	if typ != "" {
		q = q.Where("type = ?", typ)
	}
	q = q.Order("created_at DESC").Limit(limit).Offset(offset)
	return servers, q.Find(&servers).Error
}

func (r *ServerRepo) UpdateTimestamps(ctx context.Context, id string, started, stopped, terminated *time.Time) error {
	updates := map[string]interface{}{}
	if started != nil {
		updates["started_at"] = *started
	}
	if stopped != nil {
		updates["stopped_at"] = *stopped
	}
	if terminated != nil {
		updates["terminated_at"] = *terminated
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&Server{}).Where("id = ?", id).Updates(updates).Error
}
