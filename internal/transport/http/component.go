package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/HAL-X9/search-trends-service/internal/infra/config"
)

type ServerComponent struct {
	logger *slog.Logger
	server *stdhttp.Server
}

func NewServerComponent(cfg config.HTTP, handler *Handler, logger *slog.Logger) *ServerComponent {
	mux := stdhttp.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := &stdhttp.Server{
		Addr:              cfg.Addr,
		Handler:           mux,
		ReadTimeout:       cfg.Timeouts.ReadTimeout,
		ReadHeaderTimeout: cfg.Timeouts.ReadHeaderTimeout,
		WriteTimeout:      cfg.Timeouts.WriteTimeout,
		IdleTimeout:       cfg.Timeouts.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	return &ServerComponent{
		logger: logger.With("component", "http_server"),
		server: srv,
	}
}

func (s *ServerComponent) Name() string {
	return "http_server"
}

func (s *ServerComponent) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		s.logger.Info("http server started", "addr", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, stdhttp.ErrServerClosed) {
			errCh <- fmt.Errorf("http listen error: %w", err)
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.server.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}

func (s *ServerComponent) Close(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
