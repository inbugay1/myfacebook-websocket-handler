package httpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type Config struct {
	Host                          string
	Port                          string
	RequestMaxHeaderBytes         int
	ReadHeaderTimeoutMilliseconds int
}

type Server struct {
	httpServer *http.Server
}

func New(config Config, handler http.Handler) *Server {
	httpServer := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler:           handler,
		MaxHeaderBytes:    config.RequestMaxHeaderBytes,
		ReadHeaderTimeout: time.Duration(config.ReadHeaderTimeoutMilliseconds) * time.Millisecond,
	}

	return &Server{
		httpServer: httpServer,
	}
}

func (s *Server) Start() <-chan error {
	slog.Info(fmt.Sprintf("Starting HTTP server on %s", s.httpServer.Addr))

	errCh := make(chan error)

	go func() {
		err := s.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("HTTP server ListenAndServe error: %s", err)

			errCh <- fmt.Errorf("http listen and server error: %w", err)
		}
	}()

	return errCh
}

func (s *Server) Shutdown() {
	slog.Info("Shutting down the HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		slog.Error(fmt.Sprintf("HTTP server shutdown error: %s", err))

		return
	}

	slog.Info("HTTP server stopped")
}
