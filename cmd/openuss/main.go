package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	srv, err := server.Listen(ResolveAddress())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("openuss listening on %s", srv.Addr())

	if err := srv.Run(ctx); err != nil {
		log.Fatal(err)
	}

	log.Print("openuss stopped")
}
