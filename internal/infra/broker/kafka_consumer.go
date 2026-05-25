package broker

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/HAL-X9/search-trends-service/internal/infra/config"
	"github.com/HAL-X9/search-trends-service/internal/usecases"
	"github.com/twmb/franz-go/pkg/kgo"
)

type TrendsUseCase interface {
	ProcessQuery(event usecases.SearchEvent)
}

type ConsumerComponent struct {
	logger     *slog.Logger
	client     *kgo.Client
	interactor TrendsUseCase
	topic      string
	wg         sync.WaitGroup
}

func NewConsumerComponent(cfg config.KafkaConfig, interactor TrendsUseCase, logger *slog.Logger) (*ConsumerComponent, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.ConsumeTopics(cfg.Topic),
		kgo.FetchMaxBytes(5*1024*1024),
		kgo.FetchMaxWait(100*time.Millisecond),
	)
	if err != nil {
		return nil, err
	}

	return &ConsumerComponent{
		logger:     logger.With("component", "kafka_consumer"),
		client:     cl,
		interactor: interactor,
		topic:      cfg.Topic,
	}, nil
}

func (c *ConsumerComponent) Name() string {
	return "kafka_consumer"
}

func (c *ConsumerComponent) Run(ctx context.Context) error {
	c.logger.Info("kafka consumer component started, polling events", "topic", c.topic)

	c.wg.Add(1)
	defer c.wg.Done()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		fetches := c.client.PollFetches(ctx)
		if err := fetches.Err(); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			c.logger.Error("error while polling from kafka", "error", err)
			continue
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()

			var event usecases.SearchEvent
			if err := json.Unmarshal(record.Value, &event); err != nil {
				c.logger.Warn("skip invalid kafka message: failed to unmarshal SearchEvent", "error", err)
				continue
			}

			if event.Query == "" {
				continue
			}

			c.interactor.ProcessQuery(event)
		}
	}
}

func (c *ConsumerComponent) Close(ctx context.Context) error {
	c.logger.Info("shutting down kafka consumer component...")
	c.client.Close()

	c.wg.Wait()
	c.logger.Info("kafka consumer component stopped cleanly")
	return nil
}
