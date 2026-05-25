package usecases

import (
	"sync"
	"testing"
)

func TestBucket_IncrementAndReset(t *testing.T) {
	b := NewBucket()

	b.Increment("apple")
	b.Increment("apple")
	b.Increment("banana")

	b.mu.RLock()
	if b.counts["apple"] != 2 {
		t.Errorf("expected apple count to be 2, got %d", b.counts["apple"])
	}
	if b.counts["banana"] != 1 {
		t.Errorf("expected banana count to be 1, got %d", b.counts["banana"])
	}
	b.mu.RUnlock()

	b.Reset()
	b.mu.RLock()
	if len(b.counts) != 0 {
		t.Errorf("expected bucket to be empty after reset, got size %d", len(b.counts))
	}
	b.mu.RUnlock()
}

func TestSlidingWindow_AggregateAll(t *testing.T) {
	sw := NewSlidingWindow()

	sw.buckets[0].Increment("golang")
	sw.buckets[1].Increment("golang")
	sw.buckets[2].Increment("kafka")

	totals := sw.AggregateAll()

	if totals["golang"] != 2 {
		t.Errorf("expected total golang to be 2, got %d", totals["golang"])
	}
	if totals["kafka"] != 1 {
		t.Errorf("expected total kafka to be 1, got %d", totals["kafka"])
	}
}

func TestBucket_Concurrency(t *testing.T) {
	b := NewBucket()
	var wg sync.WaitGroup
	workers := 50
	iterations := 100

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				b.Increment("concurrent_word")
			}
		}()
	}

	wg.Wait()
	expected := int64(workers * iterations)
	if b.counts["concurrent_word"] != expected {
		t.Errorf("expected %d, got %d", expected, b.counts["concurrent_word"])
	}
}
