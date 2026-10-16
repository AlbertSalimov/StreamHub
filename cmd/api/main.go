package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlbertSalimov/streamhub/internal/config"
	apphttp "github.com/AlbertSalimov/streamhub/internal/http"
	"github.com/AlbertSalimov/streamhub/internal/platform"
)

func main() {
	logger := platform.NewLogger()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      apphttp.NewRouter(),
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
