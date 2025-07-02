package metrics

import (
	"context"
	"time"
)

// Repository defines the interface for metrics data persistence
type Repository interface {
	SaveTradeMetric(ctx context.Context, metric *TradeMetric) error
	CreateStrategyRun(ctx context.Context, run *StrategyRun) error
	UpdateStrategyRun(ctx context.Context, run *StrategyRun) error
	SaveTimeSeriesMetric(timestamp time.Time, symbol, metricName string, value float64, labels map[string]interface{}) error
}
