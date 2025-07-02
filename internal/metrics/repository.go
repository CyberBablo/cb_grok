package metrics

import (
	metrics_model "cb_grok/internal/models/metrics"
	"context"
	"time"
)

// Repository defines the interface for metrics data persistence
type Repository interface {
	SaveTradeMetric(ctx context.Context, metric *metrics_model.TradeMetric) error
	CreateStrategyRun(ctx context.Context, run *metrics_model.StrategyRun) error
	UpdateStrategyRun(ctx context.Context, run *metrics_model.StrategyRun) error
	SaveTimeSeriesMetric(timestamp time.Time, symbol, metricName string, value float64, labels map[string]interface{}) error
}
