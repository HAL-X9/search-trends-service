package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HAL-X9/search-trends-service/internal/usecases"
)

type mockTrendsUseCase struct {
	topTrends []usecases.WordStat
	added     string
}

func (m *mockTrendsUseCase) GetTopTrends(limit int) []usecases.WordStat { return m.topTrends }
func (m *mockTrendsUseCase) AddStopWord(word string) error              { m.added = word; return nil }
func (m *mockTrendsUseCase) RemoveStopWord(word string) error           { return nil }

func TestHandler_GetTop(t *testing.T) {
	mockUC := &mockTrendsUseCase{
		topTrends: []usecases.WordStat{
			{Word: "платье", Count: 15},
		},
	}
	handler := NewHandler(mockUC)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/trends/top?limit=1", nil)
	rr := httptest.NewRecorder()

	handler.GetTop(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var resp []WordStat
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != 1 || resp[0].Word != "платье" || resp[0].Count != 15 {
		t.Errorf("unexpected response content: %+v", resp)
	}
}

func TestHandler_AddStopWord_InvalidJSON(t *testing.T) {
	handler := NewHandler(&mockTrendsUseCase{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/stop-list", bytes.NewBufferString(`{"word": ""}`))
	rr := httptest.NewRecorder()

	handler.AddStopWord(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty word, got %d", rr.Code)
	}
}
