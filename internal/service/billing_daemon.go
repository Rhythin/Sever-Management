package service

import (
	"context"
	"time"

	"github.com/rhythin/sever-management/internal"
	"github.com/rhythin/sever-management/internal/logging"
	"github.com/rhythin/sever-management/internal/persistence"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// BillingDaemon periodically updates server billing based on uptime

type BillingDaemon struct {
	servers *persistence.ServerRepo
	cfg     *internal.Config
	db      *gorm.DB
}

func NewBillingDaemon(servers *persistence.ServerRepo, db *gorm.DB, cfg *internal.Config) *BillingDaemon {
	return &BillingDaemon{servers: servers, db: db, cfg: cfg}
}

func (b *BillingDaemon) Run(ctx context.Context) {
	zap.S().Infow("BillingDaemon started")
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			zap.S().Infow("BillingDaemon stopped")
			return
		case <-ticker.C:
			b.billAll(ctx)
		}
	}
}

func (b *BillingDaemon) billAll(ctx context.Context) {
	log := logging.S(ctx)
	log.Debugw("BillingDaemon running billAll")
	servers, err := b.servers.List(ctx, "", string("running"), "", 1000, 0)
	if err != nil {
		log.Errorw("BillingDaemon failed to list servers", "error", err)
		return
	}
	rate := b.cfg.BillingRate / 3600.0 // $/second
	now := time.Now()
	g, ctx := errgroup.WithContext(ctx)
	for _, s := range servers {
		s := s // capture loop var
		g.Go(func() error {
			if s.StartedAt == nil {
				return nil
			}
			delta := now.Sub(*s.StartedAt).Seconds()
			if delta <= 0 {
				return nil
			}
			cost := rate * delta
			err := b.db.Model(&persistence.Billing{}).Where("server_id = ?", s.ID).
				UpdateColumns(map[string]interface{}{
					"accumulated_seconds": gorm.Expr("accumulated_seconds + ?", int64(delta)),
					"total_cost":          gorm.Expr("total_cost + ?", cost),
					"last_billed_at":      now,
				}).Error
			if err != nil {
				log.Errorw("BillingDaemon failed to update billing for server", "id", s.ID, "error", err)
			} else {
				log.Infow("Billed server", "id", s.ID, "cost", cost)
			}
			return err
		})
	}
	_ = g.Wait()
}
