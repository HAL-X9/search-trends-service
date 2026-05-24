package broker

import (
	"log/slog"

	"github.com/HAL-X9/search-trends-service/internal/infra/config"
)

type ConsumerComponent struct {
	name   string
	cfg    config.KafkaConfig
	logger slog.Logger
}

func NewConsumerComponent() {

}
