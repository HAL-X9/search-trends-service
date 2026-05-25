package usecases

import (
	"log/slog"
	"sort"
	"sync"
	"time"
)

// TrendsInteractor координирует потоки данных, выполняет бизнес-правила
type TrendsInteractor struct {
	logger   *slog.Logger
	stopList StopListStorage
	window   *SlidingWindow

	topCache   []WordStat
	topCacheMu sync.RWMutex

	shutdownChan chan struct{}
}

// NewTrendsInteractor конструирует интерактор и запускает фоновый воркер агрегации
func NewTrendsInteractor(stopList StopListStorage, logger *slog.Logger) *TrendsInteractor {
	ti := &TrendsInteractor{
		logger:       logger.With("layer", "usecase"),
		stopList:     stopList,
		window:       NewSlidingWindow(),
		topCache:     make([]WordStat, 0),
		shutdownChan: make(chan struct{}),
	}

	go ti.startBackgroundAggregation()

	return ti
}

// ProcessQuery точка входа для консьюмера Кафки
func (ti *TrendsInteractor) ProcessQuery(event SearchEvent) {
	if ti.stopList.IsBanned(event.Query) {
		return
	}

	bucket := ti.window.GetCurrentBucket()

	bucket.Increment(event.Query)
}

// GetTopTrends возвращает актуальный Топ-N запросов
func (ti *TrendsInteractor) GetTopTrends(limit int) []WordStat {
	ti.topCacheMu.RLock()
	defer ti.topCacheMu.RUnlock()

	if limit > len(ti.topCache) {
		limit = len(ti.topCache)
	}
	if limit <= 0 {
		return []WordStat{}
	}

	result := make([]WordStat, limit)
	copy(result, ti.topCache[:limit])

	return result
}

// startBackgroundAggregation выделенная горутина
func (ti *TrendsInteractor) startBackgroundAggregation() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ti.shutdownChan:
			ti.logger.Info("background trends aggregator stopped")
			return

		case <-ticker.C:
			totals := ti.window.AggregateAll()

			stats := make([]WordStat, 0, len(totals))
			for word, count := range totals {
				stats = append(stats, WordStat{Word: word, Count: count})
			}

			sort.Slice(stats, func(i, j int) bool {
				if stats[i].Count == stats[j].Count {
					return stats[i].Word < stats[j].Word
				}
				return stats[i].Count > stats[j].Count
			})

			maxCachedPositions := 100
			if len(stats) > maxCachedPositions {
				stats = stats[:maxCachedPositions]
			}

			ti.topCacheMu.Lock()
			ti.topCache = stats
			ti.topCacheMu.Unlock()
		}
	}
}

func (ti *TrendsInteractor) AddStopWord(word string) error {
	ti.logger.Info("adding word to stop-list", "word", word)

	if saver, ok := ti.stopList.(interface{ Add(string) error }); ok {
		return saver.Add(word)
	}

	return nil
}

func (ti *TrendsInteractor) RemoveStopWord(word string) error {
	ti.logger.Info("removing word from stop-list", "word", word)

	if remover, ok := ti.stopList.(interface{ Remove(string) error }); ok {
		return remover.Remove(word)
	}

	return nil
}

func (ti *TrendsInteractor) Close() error {
	close(ti.shutdownChan)
	return nil
}
