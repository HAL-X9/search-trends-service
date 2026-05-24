package main

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

var words = []string{
	"платье женское", "кроссовки", "футболка", "сумка", "носки",
	"джинсы", "купальник", "серьги", "чехол на айфон", "косметика",
	"шорты", "босоножки", "рюкзак", "крем для лица", "худи", "реклама",
	"спам_запрос", "купить_дешево_акция", "тест",
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("starting search traffic generator...")

	// Инициализация Franz-Kafka
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers),
	)
	if err != nil {
		logger.Error("failed to create kafka client", "error", err)
		os.Exit(1)
	}
	defer cl.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	logger.Info("generator successfully connected to kafka, broadcasting traffic...")

	for {
		select {
		case <-ctx.Done():
			logger.Info("stopping generator gracefully...")
			return
		case <-ticker.C:
			word := words[rand.Intn(len(words))]

			record := &kgo.Record{
				Topic: "search-logs",
				Value: []byte(word),
			}

			cl.Produce(ctx, record, func(r *kgo.Record, err error) {
				if err != nil {
					logger.Error("failed to deliver message", "error", err)
				} else {
					logger.Info("message delivered successfully", "topic", r.Topic, "word", string(r.Value))
				}
			})
		}
	}
}
