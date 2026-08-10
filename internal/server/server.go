package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

const shutdownTimeout = 10 * time.Second

type Server struct {
	http     *http.Server
	listener net.Listener
}

func Listen(addr string) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		http: &http.Server{
			Handler:           http.NewServeMux(),
			ReadHeaderTimeout: 15 * time.Second,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       30 * time.Second,
		},
		listener: listener,
	}, nil
}

func (s *Server) Addr() net.Addr {
	return s.listener.Addr()
}

func (s *Server) Close() error {
	return s.http.Close()
}

func (s *Server) Run(ctx context.Context) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- s.http.Serve(s.listener) }()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	return s.http.Shutdown(shutdownCtx)
}
