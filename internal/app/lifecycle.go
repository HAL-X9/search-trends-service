package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/errgroup"
)

type Lifecycle struct {
	components []Component
	logger     *slog.Logger
}

func NewLifecycle(components []Component, logger *slog.Logger) *Lifecycle {
	return &Lifecycle{components: components, logger: logger}
}

func (l *Lifecycle) Run(ctx context.Context) error {
	eg, egCtx := errgroup.WithContext(ctx)

	for _, component := range l.components {
		c := component
		eg.Go(func() error {
			l.logger.Info("starting application component", "component", c.Name())
			if err := c.Run(egCtx); err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("component %s crashed: %w", c.Name(), err)
			}
			return nil
		})
	}

	return eg.Wait()
}

func (l *Lifecycle) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var errs error
	for _, c := range l.components {
		l.logger.Info("stopping component resource", "component", c.Name())
		if err := c.Close(ctx); err != nil {
			errs = errors.Join(errs, fmt.Errorf("failed to close %s smoothly: %w", c.Name(), err))
		}
	}

	return errs
}
