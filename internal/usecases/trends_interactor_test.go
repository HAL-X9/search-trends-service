package usecases

import (
	"log/slog"
	"os"
	"testing"
)

type mockStopList struct {
	bannedWord string
}

func (m *mockStopList) IsBanned(word string) bool { return word == m.bannedWord }
func (m *mockStopList) Add(word string) error     { return nil }
func (m *mockStopList) Remove(word string) error  { return nil }

func TestTrendsInteractor_ProcessQuery(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mockSL := &mockStopList{bannedWord: "реклама"}
	antiFraud := NewAntiFraudDetector()

	interactor := NewTrendsInteractor(mockSL, antiFraud, logger)
	defer interactor.Close()

	interactor.ProcessQuery(SearchEvent{Query: "носки"})

	interactor.ProcessQuery(SearchEvent{Query: "реклама"})

	totals := interactor.window.AggregateAll()

	if totals["носки"] != 1 {
		t.Errorf("expected count for 'носки' to be 1, got %d", totals["носки"])
	}
	if totals["реклама"] != 0 {
		t.Errorf("expected 'реклама' to be filtered out, but got count %d", totals["реклама"])
	}
}
