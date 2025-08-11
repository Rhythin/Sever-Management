package service

import (
	"context"
	"time"

	"github.com/rhythin/sever-management/internal"
	"github.com/rhythin/sever-management/internal/domain"
	"github.com/rhythin/sever-management/internal/logging"
	"github.com/rhythin/sever-management/internal/persistence"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"gorm.io/gorm"
)

// BillingDaemon periodically updates server billing based on uptime

type BillingDaemon struct {
	servers persistence.ServerRepo
	cfg     *internal.Config
	db      *gorm.DB
}

func NewBillingDaemon(servers persistence.ServerRepo, db *gorm.DB, cfg *internal.Config) *BillingDaemon {
	return &BillingDaemon{servers: servers, db: db, cfg: cfg}
}

func (b *BillingDaemon) Run(ctx context.Context) {
	interval := b.cfg.BillingInterval
	if interval <= 0 {
		interval = time.Minute
	}
	zap.S().Infow("BillingDaemon started")
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			zap.S().Infow("BillingDaemon stopped")
			return
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						zap.S().Errorw("BillingDaemon panicked; will restart loop", "recover", r)
					}
				}()
				ctx, cancel := context.WithTimeout(ctx, b.cfg.RequestTimeout)
				defer cancel()
				b.billAll(ctx)
			}()
		}
	}
}

func (b *BillingDaemon) billAll(ctx context.Context) {
	log := logging.S(ctx)
	log.Debugw("BillingDaemon running billAll")
	servers, err := b.servers.List(ctx, "", string(domain.ServerRunning), "", 1000, 0)
	if err != nil {
		log.Errorw("BillingDaemon failed to list servers", "error", err)
		return
	}
	log.Debugw("BillingDaemon found servers", "count", len(servers))
	rate := b.cfg.BillingRate / 3600.0 // $/second
	now := time.Now()
	log.Debugw("BillingDaemon billing rate", "rate", rate)
	g, gctx := errgroup.WithContext(ctx)
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
			log.Debugw("BillingDaemon billing server", "id", s.ID, "delta", delta, "cost", cost)
			err := b.servers.UpdateBilling(gctx, s.ID, int64(delta), cost)
			if err != nil {
				log.Errorw("BillingDaemon failed to update billing for server", "id", s.ID, "error", err)
			} else {
				log.Infow("Billed server", "id", s.ID, "cost", cost)
			}
			return err
		})
	}
	err = g.Wait()
	if err != nil {
		log.Errorw("BillingDaemon failed to bill servers", "error", err)
	}
}
