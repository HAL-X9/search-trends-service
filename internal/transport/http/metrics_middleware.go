package http

import (
	"net/http"
	"time"

	"github.com/HAL-X9/search-trends-service/internal/observe"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(metrics *observe.Metrics, path string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK,
		}

		next(rec, r)

		duration := time.Since(start).Seconds()
		metrics.ObserveHTTP(r.Method, path, rec.status, duration)
	}
}
