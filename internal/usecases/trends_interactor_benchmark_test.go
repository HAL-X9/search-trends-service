package usecases

import (
	"io"
	"log/slog"
	"testing"
)

type benchStopList struct{}

func (b *benchStopList) IsBanned(_ string) bool { return false }
func (b *benchStopList) Add(_ string) error     { return nil }
func (b *benchStopList) Remove(_ string) error  { return nil }

func newBenchInteractor() *TrendsInteractor {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	stopList := &benchStopList{}
	antiFraud := NewAntiFraudDetector()
	return NewTrendsInteractor(stopList, antiFraud, logger, nil)
}

func BenchmarkTrendsInteractor_ProcessQuery(b *testing.B) {
	ti := newBenchInteractor()
	defer ti.Close()

	event := SearchEvent{
		Query:     "кроссовки nike",
		UserID:    "user-1",
		IPAddress: "127.0.0.1",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ti.ProcessQuery(event)
	}
}

func BenchmarkTrendsInteractor_GetTopTrends(b *testing.B) {
	ti := newBenchInteractor()
	defer ti.Close()

	for i := 0; i < 200000; i++ {
		ti.ProcessQuery(SearchEvent{
			Query:     "товар",
			UserID:    "u",
			IPAddress: "127.0.0.1",
		})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ti.GetTopTrends(10)
	}
}

func BenchmarkSlidingWindow_AggregateAll(b *testing.B) {
	sw := NewSlidingWindow()

	for i := 0; i < 50000; i++ {
		sw.GetCurrentBucket().Increment("товар")
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = sw.AggregateAll()
	}
}
