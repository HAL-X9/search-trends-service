package broker

import (
	"log/slog"
	"os"
	"testing"

	"github.com/HAL-X9/search-trends-service/internal/usecases"
)

type mockUseCase struct{}

func (m *mockUseCase) ProcessQuery(event usecases.SearchEvent) {}

func TestConsumerComponent_Name(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockUC := &mockUseCase{}

	consumer := &ConsumerComponent{
		logger:     logger,
		interactor: mockUC,
		topic:      "test-topic",
	}

	if consumer.Name() != "kafka_consumer" {
		t.Errorf("expected component name 'kafka_consumer', got %q", consumer.Name())
	}
}
