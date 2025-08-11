package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/rhythin/sever-management/internal/logging"
	"gorm.io/gorm"
)

// ServerRepo handles server persistence

type ServerRepo struct {
	db *gorm.DB
}

func NewServerRepo(db *gorm.DB) ServerRepoInterface {
	return &ServerRepo{db: db}
}

func (r *ServerRepo) Create(ctx context.Context, s *Server) error {
	log := logging.S(ctx)
	log.Infow("ServerRepo.Create called", "id", s.ID, "region", s.Region, "type", s.Type)
	err := r.db.WithContext(ctx).Create(s).Error
	if err != nil {
		log.Errorw("ServerRepo.Create failed", "id", s.ID, "error", err)
	}
	return err
}

func (r *ServerRepo) GetByID(ctx context.Context, id string) (*Server, error) {
	log := logging.S(ctx)
	log.Debugw("ServerRepo.GetByID called", "id", id)
	var s Server
	err := r.db.WithContext(ctx).Preload("IP").Preload("Billing").Preload("Events").First(&s, "id = ?", id).Error
	if err != nil {
		log.Warnw("ServerRepo.GetByID not found or error", "id", id, "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *ServerRepo) UpdateState(ctx context.Context, id string, newState string) error {
	log := logging.S(ctx)
	log.Infow("ServerRepo.UpdateState called", "id", id, "state", newState)
	err := r.db.WithContext(ctx).Model(&Server{}).Where("id = ?", id).Update("state", newState).Error
	if err != nil {
		log.Errorw("ServerRepo.UpdateState failed", "id", id, "error", err)
	}
	return err
}

func (r *ServerRepo) List(ctx context.Context, region, status, typ string, limit, offset int) ([]*Server, error) {
	log := logging.S(ctx)
	log.Debugw("ServerRepo.List called", "region", region, "status", status, "type", typ, "limit", limit, "offset", offset)
	var servers []*Server
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
	err := q.Find(&servers).Error
	if err != nil {
		log.Errorw("ServerRepo.List failed", "error", err)
	}
	return servers, err
}

func (r *ServerRepo) UpdateTimestamps(ctx context.Context, id string, started, stopped, terminated *time.Time) error {
	log := logging.S(ctx)
	log.Debugw("ServerRepo.UpdateTimestamps called", "id", id, "started", started, "stopped", stopped, "terminated", terminated)
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
	err := r.db.WithContext(ctx).Model(&Server{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		log.Errorw("ServerRepo.UpdateTimestamps failed", "id", id, "error", err)
	}
	return err
}

func (r *ServerRepo) UpdateBilling(ctx context.Context, id string, accumulatedSeconds int64, totalCost float64) error {
	log := logging.S(ctx)
	log.Debugw("ServerRepo.UpdateBilling called", "id", id, "accumulatedSeconds", accumulatedSeconds, "totalCost", totalCost)
	err := r.db.WithContext(ctx).Model(&Billing{}).Where("server_id = ?", id).UpdateColumns(map[string]interface{}{
		"accumulated_seconds": accumulatedSeconds,
		"total_cost":          totalCost,
	}).Error
	if err != nil {
		log.Errorw("ServerRepo.UpdateBilling failed", "id", id, "error", err)
	}
	return err
}

func (r *ServerRepo) UpdateServer(ctx context.Context, id string, server *Server) error {
	log := logging.S(ctx)
	log.Debugw("ServerRepo.UpdateServer called", "id", id)
	err := r.db.WithContext(ctx).Model(&Server{}).Where("id = ?", id).Updates(server).Error
	if err != nil {
		log.Errorw("ServerRepo.UpdateServer failed", "id", id, "error", err)
	}
	return err
}
