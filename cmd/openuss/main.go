package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bwawan.com/openuss/internal/httplog"
	"bwawan.com/openuss/internal/logging"
	"bwawan.com/openuss/internal/router"
	"bwawan.com/openuss/internal/server"
	"bwawan.com/openuss/internal/tracing"
	"bwawan.com/openuss/sdk/qualifier/versioning"
	"bwawan.com/openuss/sdk/uss"
)

const FlushTimeout = 5 * time.Second

func main() {
	logger := logging.Default()
	slog.SetDefault(logger)
	err := loadDotEnv()
	if err != nil {
		logger.Error("failed to load env", slog.Any("error", err))
		os.Exit(1)
	}
	env := ParseEnvironment(os.Getenv)
	if err := run(env, logger); err != nil {
		logger.Error("openuss exited", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(env Environment, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracing, err := tracing.Init(ctx, logger)
	if err != nil {
		return err
	}
	defer HandleShutdown(shutdownTracing, logger)
	return runServer(ctx, env, logger)
}

func runServer(ctx context.Context, env Environment, logger *slog.Logger) error {
	srv, err := newServer(env, logger)
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

func newServer(env Environment, logger *slog.Logger) (*server.Server, error) {
	services, err := NewServices(env)
	if err != nil {
		return nil, err
	}
	return server.Listen(
		env.ListenAddress,
		NewHTTPHandler(services, logger),
		logger)
}

func NewHTTPHandler(services Services, logger *slog.Logger) http.Handler {
	routes := createRouter(services)
	return httplog.Middleware(logger)(routes)
}

func createRouter(services Services) http.Handler {
	return router.New(
		versioning.New(),
		services.Planning,
		uss.New(services.Intents, services.Authorizer),
	)
}

func HandleShutdown(shutdown func(context.Context) error, logger *slog.Logger) {
	flushCtx, cancel := context.WithTimeout(context.Background(), FlushTimeout)
	defer cancel()
	if err := shutdown(flushCtx); err != nil {
		logger.ErrorContext(flushCtx, "tracing shutdown failed", slog.Any("error", err))
	}
}
