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
	servers persistence.ServerRepo
	cfg     *internal.Config
}

func NewIdleReaper(servers persistence.ServerRepo, cfg *internal.Config) *IdleReaper {
	return &IdleReaper{servers: servers, cfg: cfg}
}

func (r *IdleReaper) Run(ctx context.Context) {
	interval := r.cfg.ReaperInterval
	if interval <= 0 {
		interval = time.Minute
	}
	for {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				func() {
					defer func() {
						if rec := recover(); rec != nil {
							logging.S(ctx).Errorw("IdleReaper panicked; will continue", "recover", rec)
						}
					}()
					ctx, cancel := context.WithTimeout(ctx, r.cfg.RequestTimeout)
					defer cancel()
					r.reap(ctx)
				}()
			}
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
		log.Debugw("IdleReaper checking server", "id", s.ID, "stopped_at", s.StoppedAt, "cutoff", cutoff, "now", time.Now(), "condition", s.StoppedAt.Before(cutoff))
		if s.StoppedAt != nil && s.StoppedAt.Before(cutoff) {
			s := s
			g.Go(func() error {
				if err := r.servers.UpdateState(ctx, s.ID, "terminated"); err != nil {
					log.Errorw("IdleReaper failed to terminate server", "id", s.ID, "error", err)
					return err
				}
				err = r.servers.UpdateTimestamps(ctx, s.ID, nil, nil, ptrTime(time.Now()))
				if err != nil {
					log.Errorw("IdleReaper failed to update timestamps for server", "id", s.ID, "error", err)
					return err
				}

				log.Warnw("IdleReaper terminated idle server", "id", s.ID)
				return nil
			})
		}
	}
	_ = g.Wait()
}

func ptrTime(t time.Time) *time.Time { return &t }
