package wal

import (
	"github.com/prometheus/client_golang/prometheus"

	intsegment "github.com/backbone81/write-ahead-log/internal/segment"
	intwal "github.com/backbone81/write-ahead-log/internal/wal"
)

// RegisterMetrics registers all metrics collectors with the given prometheus registerer.
func RegisterMetrics(registerer prometheus.Registerer) error {
	if err := intwal.RegisterMetrics(registerer); err != nil {
		return err
	}
	if err := intsegment.RegisterMetrics(registerer); err != nil {
		return err
	}
	return nil
}
