package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const DefaultShutdownTimeout = 10 * time.Second

type Server struct {
	HTTP            *http.Server
	ShutdownTimeout time.Duration
	listener        net.Listener
	logger          *slog.Logger
}

func Listen(addr string, handler http.Handler, logger *slog.Logger) (*Server, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Server{
		logger:          logger,
		ShutdownTimeout: DefaultShutdownTimeout,
		HTTP: &http.Server{
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
	return s.HTTP.Close()
}

func (s *Server) Run(ctx context.Context) error {
	serveErr := make(chan error, 1)
	go func() { serveErr <- s.HTTP.Serve(s.listener) }()

	select {
	case err := <-serveErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
	defer cancel()

	s.logger.InfoContext(shutdownCtx, "draining connections",
		slog.Duration("timeout", s.ShutdownTimeout))

	if err := s.HTTP.Shutdown(shutdownCtx); err != nil {
		s.logger.ErrorContext(shutdownCtx, "drain did not complete", slog.Any("error", err))
		return err
	}

	return nil
}
