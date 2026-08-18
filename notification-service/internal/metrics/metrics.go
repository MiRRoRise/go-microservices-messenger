package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var KafkaEventsConsumed = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "kafka_events_consumed_total",
		Help: "Total number of Kafka events consumed",
	},
	[]string{"topic"},
)
