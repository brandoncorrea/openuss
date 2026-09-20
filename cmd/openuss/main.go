package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/flightplanning"
	"bwawan.com/openuss/internal/httplog"
	"bwawan.com/openuss/internal/logging"
	"bwawan.com/openuss/internal/operations"
	"bwawan.com/openuss/internal/router"
	"bwawan.com/openuss/internal/server"
	"bwawan.com/openuss/internal/tracing"
	"bwawan.com/openuss/internal/versioning"
	"bwawan.com/openuss/sdk/auth"
	"bwawan.com/openuss/sdk/dss"
	"bwawan.com/openuss/sdk/peer"
	"bwawan.com/openuss/sdk/scd"
	"bwawan.com/openuss/sdk/utmclient"
	"github.com/joho/godotenv"
)

const FlushTimeout = 5 * time.Second

func ResolveAddress() string {
	port := os.Getenv("PORT")
	if port == "" {
		return ":8080"
	}
	return ":" + port
}

func loadEnv() error {
	err := godotenv.Load()
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func main() {
	logger := logging.Default()
	slog.SetDefault(logger)
	err := loadEnv()
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
	defer HandleShutdown(shutdownTracing, logger)
	return runServer(ctx, logger)
}

func runServer(ctx context.Context, logger *slog.Logger) error {
	srv, err := newServer(logger)
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

func newServer(logger *slog.Logger) (*server.Server, error) {
	tokens, err := NewTokenSource()
	if err != nil {
		return nil, err
	}
	intents := scd.NewInMemoryIntentStore()
	planning, err := NewPlanningHandler(tokens, intents)
	if err != nil {
		return nil, err
	}
	return server.Listen(
		ResolveAddress(),
		NewHTTPHandler(intents, planning, logger),
		logger)
}

func NewTokenSource() (auth.TokenSource, error) {
	if os.Getenv("TOKEN_IMPL") == "memory" {
		return auth.NewInMemoryTokenSource(), nil
	}
	endpoint := os.Getenv("OAUTH_ENDPOINT")
	sub := os.Getenv("OAUTH_SUB")
	return auth.NewDummyOAuth(endpoint, sub, nil)
}

func NewHTTPHandler(intents scd.IntentStore, planning router.FlightPlanning, logger *slog.Logger) http.Handler {
	routes := createRouter(intents, planning)
	return httplog.Middleware(logger)(routes)
}

func createRouter(intents scd.IntentStore, planning router.FlightPlanning) http.Handler {
	return router.New(
		versioning.New(),
		planning,
		operations.New(intents),
	)
}

func NewPlanningHandler(tokens auth.TokenSource, intents scd.IntentStore) (*flightplanning.Handler, error) {
	ussBaseURL := os.Getenv("USS_BASE_URL")
	if strings.TrimSpace(ussBaseURL) == "" {
		return nil, errors.New("USS_BASE_URL is required")
	}

	client := utmclient.New(tokens, nil)
	dssClient := newDSSClient(client)
	peer := peer.New(client)
	flights := db.NewInMemoryFlightStore()
	service := scd.New(dssClient, peer, intents, ussBaseURL)
	return flightplanning.New(service, flights), nil
}

func newDSSClient(client *utmclient.Client) dss.Client {
	if os.Getenv("DSS_IMPL") == "memory" {
		return dss.NewInMemoryDSS()
	}
	return dss.New(os.Getenv("DSS_BASE_URL"), client)
}

func HandleShutdown(shutdown func(context.Context) error, logger *slog.Logger) {
	flushCtx, cancel := context.WithTimeout(context.Background(), FlushTimeout)
	defer cancel()
	if err := shutdown(flushCtx); err != nil {
		logger.ErrorContext(flushCtx, "tracing shutdown failed", slog.Any("error", err))
	}
}
