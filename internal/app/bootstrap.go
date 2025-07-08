package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kv-storage/internal/config"
	"kv-storage/internal/interfaces"
	"kv-storage/internal/repository"
	"kv-storage/internal/service"
	"kv-storage/internal/transport/http"
)

type Application struct {
	router interfaces.Router
	logger interfaces.Logger
	config *config.Config
	repo   interfaces.KVRepository
}

func Bootstrap() (*Application, error) {
	cfg := config.MustLoad("config/config.yaml")

	logger := NewLogger(cfg.App.Environment)

	repo, err := repository.NewTarantoolRepository(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize repository: %w", err)
	}

	kvService := service.NewKVService(repo, logger)

	router := http.NewRouter(cfg, logger, kvService)

	return &Application{
		router: router,
		logger: logger,
		config: cfg,
		repo:   repo,
	}, nil
}

func (a *Application) Run() {
	a.logger.Info("Starting KV Storage application",
		"port", a.config.HTTPServer.Port,
		"environment", a.config.App.Environment,
	)

	go func() {
		if err := a.router.Run(":" + a.config.HTTPServer.Port); err != nil {
			a.logger.Error("HTTP server error", "error", err)
		}
	}()

	a.waitForShutdown()
}

func (a *Application) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.router.Shutdown(ctx); err != nil {
		a.logger.Error("Error during server shutdown", "error", err)
	}

	if err := a.repo.Close(); err != nil {
		a.logger.Error("Error closing repository", "error", err)
	}

	a.logger.Info("Application shutdown completed")
}
