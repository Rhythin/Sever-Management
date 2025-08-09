package service

import (
	"context"
	"time"

	"github.com/rhythin/sever-management/internal"
	"github.com/rhythin/sever-management/internal/logging"
	"github.com/rhythin/sever-management/internal/persistence"
	"golang.org/x/sync/errgroup"
)

// IdleReaper terminates servers stopped for longer than IdleTimeout

type IdleReaper struct {
	servers *persistence.ServerRepo
	cfg     *internal.Config
}

func NewIdleReaper(servers *persistence.ServerRepo, cfg *internal.Config) *IdleReaper {
	return &IdleReaper{servers: servers, cfg: cfg}
}

func (r *IdleReaper) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute) // runs every minute
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.reap(ctx)
		}
	}
}

func (r *IdleReaper) reap(ctx context.Context) {
	log := logging.S(ctx)
	log.Debugw("IdleReaper running reap")
	servers, err := r.servers.List(ctx, "", string("stopped"), "", 1000, 0)
	if err != nil {
		log.Errorw("IdleReaper failed to list servers", "error", err)
		return
	}
	cutoff := time.Now().Add(-r.cfg.IdleTimeout)
	g, ctx := errgroup.WithContext(ctx)
	for _, s := range servers {
		s := s // capture loop var
		if s.StoppedAt != nil && s.StoppedAt.Before(cutoff) {
			s := s
			g.Go(func() error {
				if err := r.servers.UpdateState(ctx, s.ID, "terminated"); err != nil {
					log.Errorw("IdleReaper failed to terminate server", "id", s.ID, "error", err)
					return err
				}
				_ = r.servers.UpdateTimestamps(ctx, s.ID, nil, nil, ptrTime(time.Now()))
				log.Warnw("IdleReaper terminated idle server", "id", s.ID)
				return nil
			})
		}
	}
	_ = g.Wait()
}

func ptrTime(t time.Time) *time.Time { return &t }
