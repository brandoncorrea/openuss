package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bwawan.com/openuss/internal/api/scdussv1"
	"bwawan.com/openuss/internal/auth"
	"bwawan.com/openuss/internal/db"
	"bwawan.com/openuss/internal/dss"
	"bwawan.com/openuss/internal/flightplanning"
	"bwawan.com/openuss/internal/httplog"
	"bwawan.com/openuss/internal/logging"
	"bwawan.com/openuss/internal/operations"
	"bwawan.com/openuss/internal/router"
	"bwawan.com/openuss/internal/server"
	"bwawan.com/openuss/internal/tracing"
	"bwawan.com/openuss/internal/util"
	"bwawan.com/openuss/internal/utmclient"
	"bwawan.com/openuss/internal/versioning"
	"github.com/joho/godotenv"
)

const flushTimeout = 5 * time.Second

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
	defer handleShutdown(shutdownTracing, logger)
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
	auth, err := newTokenSource()
	if err != nil {
		return nil, err
	}
	db := db.NewInMemoryDB()
	planning, err := newPlanningHandler(auth, db)
	if err != nil {
		return nil, err
	}
	return server.Listen(
		ResolveAddress(),
		newHandler(db, planning, logger),
		logger)
}

func newTokenSource() (auth.TokenSource, error) {
	if os.Getenv("TOKEN_IMPL") == "memory" {
		return auth.NewInMemoryTokenSource(), nil
	}
	endpoint := os.Getenv("OAUTH_ENDPOINT")
	sub := os.Getenv("OAUTH_SUB")
	return auth.NewDummyOAuth(endpoint, sub, nil)
}

func newHandler(db db.DB, planning router.FlightPlanning, logger *slog.Logger) http.Handler {
	routes := createRouter(db, planning)
	return httplog.Middleware(logger)(routes)
}

func createRouter(db db.DB, planning router.FlightPlanning) http.Handler {
	return router.New(
		&versioning.Handler{},
		planning,
		&operations.Handler{DB: db},
	)
}

func newPlanningHandler(tokenSource auth.TokenSource, db db.DB) (*flightplanning.Handler, error) {
	ussBaseUrl := os.Getenv("USS_BASE_URL")
	if util.IsBlank(ussBaseUrl) {
		return nil, errors.New("USS_BASE_URL is required")
	}
	return &flightplanning.Handler{
		DSS:        newUssAuthority(tokenSource),
		DB:         db,
		UssBaseUrl: scdussv1.OperationalIntentUssBaseURL(ussBaseUrl),
	}, nil
}

func newUssAuthority(tokenSource auth.TokenSource) dss.USSAuthority {
	if os.Getenv("DSS_IMPL") == "memory" {
		return dss.NewInMemoryDSS()
	}
	return newRealDss(tokenSource)
}

func newRealDss(tokenSource auth.TokenSource) dss.USSAuthority {
	return &dss.DSS{
		Host:   os.Getenv("DSS_BASE_URL"),
		Client: utmclient.New(tokenSource, nil),
	}
}

func handleShutdown(shutdown func(context.Context) error, logger *slog.Logger) {
	flushCtx, cancel := context.WithTimeout(context.Background(), flushTimeout)
	defer cancel()
	if err := shutdown(flushCtx); err != nil {
		logger.ErrorContext(flushCtx, "tracing shutdown failed", slog.Any("error", err))
	}
}
