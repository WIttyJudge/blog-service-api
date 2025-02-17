package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var cacheMetrics = promauto.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "cache",
		Help: "Metrics of hits and misses",
	}, []string{"cache_name", "result"},
)
