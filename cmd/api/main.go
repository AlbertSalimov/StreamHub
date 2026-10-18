package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	appcache "github.com/AlbertSalimov/streamhub/internal/cache"
	"github.com/AlbertSalimov/streamhub/internal/config"
	"github.com/AlbertSalimov/streamhub/internal/events"
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

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB := 0
	if raw := os.Getenv("REDIS_DB"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 {
			slog.Error("invalid REDIS_DB")
			return
		}
		redisDB = value
	}

	redisCache := appcache.NewRedis(redisAddr, redisPassword, redisDB)
	defer redisCache.Close()

	pingCtx, cancelPing := context.WithTimeout(context.Background(), time.Second)
	err = redisCache.Ping(pingCtx)
	cancelPing()

	if err != nil {
		slog.Warn("Redis unavailable; cache will be bypassed on errors", "error", err)
	}

	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	natsCtx, cancelNATS := context.WithTimeout(context.Background(), 5*time.Second)
	publisher, err := events.NewPublisher(natsCtx, natsURL)
	cancelNATS()

	if err != nil {
		slog.Error("failed to initialize NATS publisher", "error", err)
		return
	}
	defer publisher.Close()

	server := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      apphttp.NewRouter(db, redisCache, publisher),
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

	go func() {
		if err := events.RunConsumer(ctx, natsURL); err != nil && ctx.Err() == nil {
			slog.Error("NATS consumer stopped", "error", err)
		}
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
