package http

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/trends/top", h.GetTop)

	mux.HandleFunc("POST /api/v1/stop-list", h.AddStopWord)
	mux.HandleFunc("DELETE /api/v1/stop-list", h.RemoveStopWord)
}
