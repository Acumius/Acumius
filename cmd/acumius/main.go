package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Acumius/Acumius/internal/api"
	"github.com/Acumius/Acumius/internal/config"
	"github.com/Acumius/Acumius/internal/storage"
)

func main() {
	logger := slog.With("service", "acumius")

	cfg := config.Load()

	pg, err := storage.NewPostgresStore(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
	} else {
		defer pg.Close()
	}

	vk, err := storage.NewValkeyStore(cfg.ValkeyURL)
	if err != nil {
		logger.Error("failed to connect to valkey", "error", err)
	} else {
		defer vk.Close()
	}

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           api.NewMux(pg, vk),
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", "port", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
		return
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
