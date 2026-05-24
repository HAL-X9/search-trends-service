package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/HAL-X9/search-trends-service/internal/infra/broker"
	"github.com/HAL-X9/search-trends-service/internal/infra/config"
)

// Component описывает абстрактный инфраструктурный контракт для запускаемых сервисов
type Component interface {
	Name() string
	Run(ctx context.Context) error
	Close(ctx context.Context) error
}

// App корень композиции приложения
type App struct {
	runtime *Runtime
}

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "", "path to app config (overrides env)")
	flag.Parse()
}

// Run инициализирует и запускает приложение. Вызывается из main.go
func Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if configPath == "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		logger.Error("failed to load configuration:", "error", err)
		return err
	}

	application, err := New(cfg, logger)
	if err != nil {
		logger.Error("failed to bootstrap application:", "error", err)
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

type mockUseCase struct {
	logger *slog.Logger
}

func (m *mockUseCase) ProcessQuery(ctx context.Context, query string) {
	m.logger.Info("лог до usecase", "query", query)
}

// New производит DI (Dependency Injection) сборку всех слоев приложения
func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	mockMock := &mockUseCase{logger: logger}

	kafkaConsumer, err := broker.NewConsumerComponent(cfg.KafkaConfig, mockMock, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init kafka consumer: %w", err)
	}

	components := []Component{kafkaConsumer}

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
