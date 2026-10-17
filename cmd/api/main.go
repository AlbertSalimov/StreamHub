package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlbertSalimov/streamhub/internal/config"
	apphttp "github.com/AlbertSalimov/streamhub/internal/http"
	"github.com/AlbertSalimov/streamhub/internal/platform"
	"github.com/AlbertSalimov/streamhub/internal/postgres"
)

func main() {
	logger := platform.NewLogger()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	startupCtx, cancelStartup := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)

	db, err := postgres.NewPool(startupCtx, cfg.DatabaseURL)
	if err != nil {
		cancelStartup()
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
		cancelStartup()
		db.Close()
		logger.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	cancelStartup()

	defer db.Close()

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      apphttp.NewRouter(db),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("server started", "address", cfg.HTTPAddr)

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
		close(serverErrors)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")

	case err := <-serverErrors:
		if err != nil {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)

		if closeErr := server.Close(); closeErr != nil {
			logger.Error("forced close failed", "error", closeErr)
		}

		os.Exit(1)
	}

	logger.Info("server stopped")
}
