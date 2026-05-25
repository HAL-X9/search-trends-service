package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/HAL-X9/search-trends-service/internal/observe"
	"github.com/HAL-X9/search-trends-service/internal/usecases"
)

type WordStat struct {
	Word  string `json:"word"`
	Count int64  `json:"count"`
}

type StopListRequest struct {
	Word string `json:"word"`
}

type TrendsUseCase interface {
	GetTopTrends(limit int) []usecases.WordStat
	AddStopWord(word string) error
	RemoveStopWord(word string) error
}

type Handler struct {
	interactor TrendsUseCase
	metrics    *observe.Metrics
}

func NewHandler(interactor TrendsUseCase, metrics *observe.Metrics) *Handler {
	return &Handler{
		interactor: interactor,
		metrics:    metrics,
	}
}

// GetTop обрабатывает запрос на получение Топ-N популярных запросов за 5 минут
func (h *Handler) GetTop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limitParam := r.URL.Query().Get("limit")

	limit := 10 // По умолчанию
	if limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	domainTop := h.interactor.GetTopTrends(limit)

	response := make([]WordStat, len(domainTop))
	for i, v := range domainTop {
		response[i] = WordStat{
			Word:  v.Word,
			Count: v.Count,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// AddStopWord добавляет нежелательное слово в бан-лист
func (h *Handler) AddStopWord(w http.ResponseWriter, r *http.Request) {
	var req StopListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Word == "" {
		http.Error(w, `{"error":"invalid_json_or_empty_word"}`, http.StatusBadRequest)
		return
	}

	if err := h.interactor.AddStopWord(req.Word); err != nil {
		http.Error(w, `{"error":"failed_to_add_word"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"success","message":"word_banned"}`))
}

// RemoveStopWord удаляет слово из бан-листа
func (h *Handler) RemoveStopWord(w http.ResponseWriter, r *http.Request) {
	var req StopListRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Word == "" {
		http.Error(w, `{"error":"invalid_json_or_empty_word"}`, http.StatusBadRequest)
		return
	}

	if err := h.interactor.RemoveStopWord(req.Word); err != nil {
		http.Error(w, `{"error":"failed_to_remove_word"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"success","message":"word_unbanned"}`))
}
