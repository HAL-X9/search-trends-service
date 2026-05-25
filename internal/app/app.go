package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	consumer "github.com/HAL-X9/search-trends-service/internal/infra/broker"
	"github.com/HAL-X9/search-trends-service/internal/infra/config"
	repo "github.com/HAL-X9/search-trends-service/internal/repository"
	httptransport "github.com/HAL-X9/search-trends-service/internal/transport/http"
	"github.com/HAL-X9/search-trends-service/internal/usecases"
)

type Component interface {
	Name() string
	Run(ctx context.Context) error
	Close(ctx context.Context) error
}

type App struct {
	runtime *Runtime
}

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "path to app config (overrides env)")
	flag.Parse()
}

func Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		return err
	}

	application, err := New(cfg, logger)
	if err != nil {
		logger.Error("failed to bootstrap application", "error", err)
		return err
	}

	defer func() {
		if closeErr := application.Close(); closeErr != nil {
			logger.Error("error during application closure", "error", closeErr)
		}
	}()

	if err = application.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("application runtime failure", "error", err)
		return err
	}

	logger.Info("application stopped successfully")
	return nil
}

func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	stopListStorage, err := repo.NewStopListStorage("config/stop-list.txt")
	if err != nil {
		return nil, fmt.Errorf("failed to init stop-list storage: %w", err)
	}

	antiFraud := usecases.NewAntiFraudDetector()

	trendsInteractor := usecases.NewTrendsInteractor(stopListStorage, antiFraud, logger)

	httpHandler := httptransport.NewHandler(trendsInteractor)
	httpComponent := httptransport.NewServerComponent(cfg.HTTP, httpHandler, logger)

	kafkaConsumer, err := consumer.NewConsumerComponent(cfg.KafkaConfig, trendsInteractor, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init kafka consumer: %w", err)
	}

	components := []Component{
		httpComponent,
		kafkaConsumer,
	}

	lifecycle := NewLifecycle(components, logger)
	runtime := NewRuntime(lifecycle, 15*time.Second, logger)

	return &App{runtime: runtime}, nil
}

func (a *App) Run(ctx context.Context) error {
	return a.runtime.Run(ctx)
}

func (a *App) Close() error {
	return a.runtime.Close()
}
