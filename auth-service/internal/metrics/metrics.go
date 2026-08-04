package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_http_requests_total",
			Help: "Total number of http requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "auth_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint"},
	)

	RegistrationTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_registrations_total",
			Help: "Total number of user registrations",
		},
	)

	LoginsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "auth_logins_total",
			Help: "Total number of user logins",
		},
	)

	ActiveUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "auth_active_users",
			Help: "Number of active users",
		},
	)
)

func RecordRequest(method, endpoint string, statusCode int64, duration float64) {
	RequestsTotal.WithLabelValues(method, endpoint, strconv.FormatInt(statusCode, 10)).Inc()
	RequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}
