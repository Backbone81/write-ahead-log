package wal

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ReadEntryTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "wal_read_entry_total",
			Help: "Total number of entries read.",
		},
	)
	ReadEntryBytes = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "wal_read_entry_bytes_total",
			Help: "Total number of bytes read (excluding metadata).",
		},
	)

	AppendEntryTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "wal_append_entry_total",
			Help: "Total number of entries appended.",
		},
	)
	AppendEntryBytes = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "wal_append_entry_bytes_total",
			Help: "Total number of bytes appended (excluding metadata).",
		},
	)

	SyncTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "wal_sync_total",
			Help: "Total number of syncs executed.",
		},
	)
	SyncDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "wal_sync_duration_seconds",
			Help:    "Duration of syncs in seconds.",
			Buckets: prometheus.ExponentialBuckets(0.0001, 2, 16),
		},
	)

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
		ReadEntryTotal,
		ReadEntryBytes,

		AppendEntryTotal,
		AppendEntryBytes,

		SyncTotal,
		SyncDuration,

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
