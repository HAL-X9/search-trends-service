package usecases

import (
	"log/slog"
	"sort"
	"sync"
	"time"
)

type TrendsInteractor struct {
	logger    *slog.Logger
	stopList  StopListStorage
	antiFraud *AntiFraudDetector
	window    *SlidingWindow

	topCache   []WordStat
	topCacheMu sync.RWMutex

	shutdownChan chan struct{}
}

func NewTrendsInteractor(stopList StopListStorage, antiFraud *AntiFraudDetector, logger *slog.Logger) *TrendsInteractor {
	ti := &TrendsInteractor{
		logger:       logger.With("layer", "usecase"),
		stopList:     stopList,
		antiFraud:    antiFraud,
		window:       NewSlidingWindow(),
		topCache:     make([]WordStat, 0),
		shutdownChan: make(chan struct{}),
	}

	go ti.startBackgroundAggregation()
	return ti
}

func (ti *TrendsInteractor) ProcessQuery(event SearchEvent) {
	if event.Query == "" {
		return
	}

	if ti.stopList.IsBanned(event.Query) {
		ti.logger.Debug("query dropped by stop-list", "query", event.Query)
		return
	}

	if ti.antiFraud != nil && ti.antiFraud.IsSpam(event) {
		return
	}

	bucket := ti.window.GetCurrentBucket()
	bucket.Increment(event.Query)
}

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
	return ti.stopList.Add(word)
}

func (ti *TrendsInteractor) RemoveStopWord(word string) error {
	return ti.stopList.Remove(word)
}

func (ti *TrendsInteractor) Close() error {
	close(ti.shutdownChan)
	return nil
}
