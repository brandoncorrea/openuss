package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const defaultShutdownTimeout = 10 * time.Second

type Server struct {
	http            *http.Server
	listener        net.Listener
	logger          *slog.Logger
	shutdownTimeout time.Duration
}

func Listen(addr string, handler http.Handler, logger *slog.Logger) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:          logger,
		shutdownTimeout: defaultShutdownTimeout,
		http: &http.Server{
			Handler:           handler,
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	s.logger.InfoContext(shutdownCtx, "draining connections",
		slog.Duration("timeout", s.shutdownTimeout))

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		s.logger.ErrorContext(shutdownCtx, "drain did not complete", slog.Any("error", err))
		return err
	}

	return nil
}
