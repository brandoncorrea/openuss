package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bwawan.com/openuss/internal/flightplanning"
	"bwawan.com/openuss/internal/logging"
	"bwawan.com/openuss/internal/router"
	"bwawan.com/openuss/internal/server"
	"bwawan.com/openuss/internal/tracing"
	"bwawan.com/openuss/internal/versioning"
	"github.com/joho/godotenv"
)

const flushTimeout = 5 * time.Second

func ResolveAddress() string {
	port := os.Getenv("PORT")
	if port == "" {
		return ":80"
	}
	return ":" + port
}

func main() {
	logger := logging.Default()
	err := godotenv.Load()
	if err != nil {
		logger.Error("failed to load env", slog.Any("error", err))
		os.Exit(1)
	}
	if err := run(logger); err != nil {
		logger.Error("openuss exited", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := tracing.Init(ctx, logger)
	if err != nil {
		return err
	}
	defer handleShutdown(shutdownTracing, logger)

	handler := router.New(&versioning.Handler{}, &flightplanning.Handler{})
	srv, err := server.Listen(ResolveAddress(), handler, logger)
	if err != nil {
		return err
	}

	logger.InfoContext(ctx, "openuss listening", slog.String("addr", srv.Addr().String()))

	if err := srv.Run(ctx); err != nil {
		return err
	}

	logger.InfoContext(ctx, "openuss stopped")

	return nil
}

func handleShutdown(shutdown func(context.Context) error, logger *slog.Logger) {
	flushCtx, cancel := context.WithTimeout(context.Background(), flushTimeout)
	defer cancel()
	if err := shutdown(flushCtx); err != nil {
		logger.ErrorContext(flushCtx, "tracing shutdown failed", slog.Any("error", err))
	}
}
