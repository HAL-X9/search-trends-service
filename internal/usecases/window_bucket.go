package usecases

import (
	"sync"
	"time"
)

// Bucket инкапсулирует метрики за одну секунду
type Bucket struct {
	mu     sync.RWMutex
	counts map[string]int64
}

func NewBucket() *Bucket {
	return &Bucket{
		counts: make(map[string]int64),
	}
}

// Increment увеличивает счетчик на +1
func (b *Bucket) Increment(word string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.counts[word]++
}

// Reset очищает бакет для повторного использования
func (b *Bucket) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for key := range b.counts {
		delete(b.counts, key)
	}
}

// SlidingWindow управляет массивом бакетов и реализует логику скользящего окна
type SlidingWindow struct {
	mu          sync.RWMutex
	buckets     []*Bucket
	size        int
	cursor      int
	lastShifted time.Time
}

func NewSlidingWindow() *SlidingWindow {
	windowSize := 300 // 5 минут
	buckets := make([]*Bucket, windowSize)
	for i := 0; i < windowSize; i++ {
		buckets[i] = NewBucket()
	}

	return &SlidingWindow{
		buckets:     buckets,
		size:        windowSize,
		cursor:      0,
		lastShifted: time.Now(),
	}
}

// GetCurrentBucket определяет бакет текущей секунды и сдвигает окно (очищая старье)
func (sw *SlidingWindow) GetCurrentBucket() *Bucket {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	elapsed := int(now.Sub(sw.lastShifted).Seconds())

	if elapsed > 0 {
		if elapsed > sw.size {
			elapsed = sw.size
		}

		for i := 0; i < elapsed; i++ {
			sw.cursor = (sw.cursor + 1) % sw.size
			sw.buckets[sw.cursor].Reset()
		}

		sw.lastShifted = now
	}

	return sw.buckets[sw.cursor]
}

// AggregateAll собирает общую статистику по всем 300 бакетам
func (sw *SlidingWindow) AggregateAll() map[string]int64 {
	sw.mu.RLock()
	defer sw.mu.RUnlock()

	totals := make(map[string]int64)

	for _, bucket := range sw.buckets {
		bucket.mu.RLock()
		for word, count := range bucket.counts {
			totals[word] += count
		}
		bucket.mu.RUnlock()
	}

	return totals
}
