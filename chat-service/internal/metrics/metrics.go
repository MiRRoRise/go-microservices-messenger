package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "endpoint", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"service", "method", "endpoint"},
	)

	MessagesCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "message_created_total",
			Help: "Total number of messages created",
		},
	)
)

const serviceName = "chat"

func RecordRequest(method, endpoint string, statusCode int64, duration float64) {
	RequestsTotal.WithLabelValues(serviceName, method, endpoint, strconv.FormatInt(statusCode, 10)).Inc()
	RequestDuration.WithLabelValues(serviceName, method, endpoint).Observe(duration)
}
