package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Runtime struct {
	lifecycle       *Lifecycle
	shutdownTimeout time.Duration
	logger          *slog.Logger
}

func NewRuntime(lc *Lifecycle, timeout time.Duration, logger *slog.Logger) *Runtime {
	return &Runtime{lifecycle: lc, shutdownTimeout: timeout, logger: logger}
}

func (r *Runtime) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	errCh := make(chan error, 1)
	go func() {
		errCh <- r.lifecycle.Run(runCtx)
	}()

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		r.logger.Info("OS signal received. Initiating graceful shutdown...", "signal", sig.String())
		cancel()

		select {
		case <-errCh:
			return r.Close()
		case <-time.After(r.shutdownTimeout):
			return fmt.Errorf("graceful shutdown forced to stop due to timeout: %v", r.shutdownTimeout)
		}
	}
}

func (r *Runtime) Close() error {
	return r.lifecycle.Close()
}
