package observe

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	KafkaEvents         *prometheus.CounterVec
	DroppedEvents       *prometheus.CounterVec
	HTTPRequestTotal    *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	AppUp               prometheus.Gauge
}

func NewMetrics() *Metrics {
	m := &Metrics{
		KafkaEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "search_trends",
				Name:      "kafka_events_total",
				Help:      "Total number of kafka events consumed",
			},
			[]string{"status"},
		),
		DroppedEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "search_trends",
				Name:      "dropped_events_total",
				Help:      "Total number of dropped events",
			},
			[]string{"reason"},
		),
		HTTPRequestTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "search_trends",
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "path", "status"},
		),
		HTTPRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "search_trends",
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request latency in seconds",
				Buckets:   []float64{0.001, 0.003, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
			},
			[]string{"method", "path"},
		),
		AppUp: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: "search_trends",
				Name:      "app_up",
				Help:      "Is app alive (1 = up)",
			},
		),
	}
	prometheus.MustRegister(
		m.KafkaEvents,
		m.DroppedEvents,
		m.HTTPRequestTotal,
		m.HTTPRequestDuration,
		m.AppUp,
	)

	m.AppUp.Set(1)
	return m
}

func (m *Metrics) ObserveHTTP(method, path string, statusCode int, seconds float64) {
	m.HTTPRequestTotal.WithLabelValues(method, path, strconv.Itoa(statusCode)).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, path).Observe(seconds)
}
