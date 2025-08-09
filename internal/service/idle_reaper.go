package service

import (
	"context"
	"time"

	"github.com/rhythin/sever-management/internal"
	"github.com/rhythin/sever-management/internal/persistence"
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
	servers, err := r.servers.List(ctx, "", string("stopped"), "", 1000, 0)
	if err != nil {
		return
	}
	cutoff := time.Now().Add(-r.cfg.IdleTimeout)
	for _, s := range servers {
		if s.StoppedAt != nil && s.StoppedAt.Before(cutoff) {
			_ = r.servers.UpdateState(ctx, s.ID, "terminated")
			_ = r.servers.UpdateTimestamps(ctx, s.ID, nil, nil, ptrTime(time.Now()))
		}
	}
}

func ptrTime(t time.Time) *time.Time { return &t }
