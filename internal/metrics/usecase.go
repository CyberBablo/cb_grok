package metrics

import "cb_grok/internal/trader"

// Collector defines the interface for metrics collection
type Collector interface {
	SaveIndicatorData(timestamp int64, indicators map[string]float64) error
	SaveTradeMetric(order trader.Action, indicators map[string]float64) error
	Close() error
}
