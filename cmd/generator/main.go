package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HAL-X9/search-trends-service/internal/usecases"
	"github.com/twmb/franz-go/pkg/kgo"
)

var words = []string{
	"платье женское", "кроссовки", "футболка", "сумка", "носки",
	"джинсы", "купальник", "серьги", "чехол на айфон", "косметика",
	"шорты", "рюкзак", "крем для лица", "худи", "реклама",
	"спам_запрос", "купить_дешево_акция", "тест",
	"наушники беспроводные", "смарт-часы", "пауэрбанк", "плед пушистый", "термокружка",
	"увлажнитель воздуха", "солнцезащитные очки",
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("starting search traffic generator...")

	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "search-logs"
	}

	cl, err := kgo.NewClient(kgo.SeedBrokers(brokers))
	if err != nil {
		logger.Error("failed to create kafka client", "error", err)
		os.Exit(1)
	}
	defer cl.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("stopping generator gracefully...")
			return
		case <-ticker.C:
			event := usecases.SearchEvent{
				Query:     words[rand.Intn(len(words))],
				UserID:    "usr_demo_" + randomSuffix(4),
				IPAddress: "192.168.1." + randomSuffix(2),
				Timestamp: time.Now().Unix(),
			}

			payload, err := json.Marshal(event)
			if err != nil {
				logger.Error("failed to marshal event", "error", err)
				continue
			}

			record := &kgo.Record{
				Topic: topic,
				Value: payload,
			}

			cl.Produce(ctx, record, func(r *kgo.Record, err error) {
				if err != nil {
					logger.Error("failed to deliver message", "topic", topic, "error", err)
					return
				}
				logger.Info("message delivered", "topic", r.Topic)
			})
		}
	}
}

func randomSuffix(n int) string {
	const letters = "0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
