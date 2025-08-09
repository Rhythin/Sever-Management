package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rhythin/sever-management/internal"
	"github.com/rhythin/sever-management/internal/api"
	"github.com/rhythin/sever-management/internal/logging"
	"github.com/rhythin/sever-management/internal/persistence"
	"github.com/rhythin/sever-management/internal/service"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			internal.LoadConfig,
			persistence.NewDB,
			persistence.NewServerRepo,
			persistence.NewIPRepo,
			persistence.NewEventRepo,
			service.NewServerService,
			service.NewBillingDaemon,
			service.NewIdleReaper,
			logging.InitLogger,
			func(svc *service.ServerService, repo *persistence.ServerRepo) *api.ServerHandlers {
				return &api.ServerHandlers{Service: svc, Repo: repo}
			},
			api.NewRouter,
		),
		fx.Invoke(runServer),
	).Run()
}

func runServer(
	lc fx.Lifecycle,
	cfg *internal.Config,
	r http.Handler,
	billing *service.BillingDaemon,
	reaper *service.IdleReaper,
	logger *zap.Logger,
) {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler: r,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			zap.S().Infof("Starting server on :%d", cfg.HTTPPort)
			go billing.Run(ctx)
			go reaper.Run(ctx)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					zap.S().Errorw("HTTP server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			zap.S().Infow("Shutting down HTTP server...")
			return server.Shutdown(ctx)
		},
	})

	// Graceful shutdown on SIGINT/SIGTERM
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c
		zap.S().Warnw("Received shutdown signal, shutting down...")
		server.Shutdown(context.Background())
		os.Exit(0)
	}()
}
