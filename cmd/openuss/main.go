package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"bwawan.com/openuss/internal/logging"
	"bwawan.com/openuss/internal/server"
)

func ResolveAddress() string {
	port := os.Getenv("PORT")
	if port == "" {
		return ":80"
	}
	return ":" + port
}

func main() {
	logger := logging.Default()
	if err := run(logger); err != nil {
		logger.Error("openuss exited", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv, err := server.Listen(ResolveAddress(), http.NewServeMux(), logger)
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
