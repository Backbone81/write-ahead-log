package wal

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	RolloverTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "wal_rollover_total",
			Help: "Total number of rollovers executed.",
		},
	)

	RolloverDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "wal_rollover_duration_seconds",
			Help:    "Duration of rollovers in seconds.",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 16),
		},
	)
)

// RegisterMetrics registers all metrics collectors with the given prometheus registerer.
func RegisterMetrics(registerer prometheus.Registerer) error {
	metrics := []prometheus.Collector{
		RolloverTotal,
		RolloverDuration,
	}
	for _, metric := range metrics {
		if err := registerer.Register(metric); err != nil {
			return err
		}
	}
	return nil
}
