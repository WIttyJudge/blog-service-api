package api

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "http",
		Name:      "requests_total",
		Help:      "Total number of requests processed",
	}, []string{"status"})

	httpRequestDurationSec = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "http",
		Name:      "request_duration",
		Help:      "Duration of HTTP requests in seconds",
		Buckets:   prometheus.DefBuckets,
	})

	httpCurrentDurationSec = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "http",
		Name:      "current_requests",
		Help:      "Current number of requests in process",
	})
)
