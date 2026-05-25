package http

import "net/http"

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/trends/top", h.GetTop)
	mux.HandleFunc("/api/v1/stop-list", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			h.AddStopWord(w, r)
		case http.MethodDelete:
			h.RemoveStopWord(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}
