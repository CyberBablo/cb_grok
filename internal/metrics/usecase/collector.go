package usecase

import (
	"cb_grok/internal/metrics"
	"cb_grok/internal/metrics/repository"
	"cb_grok/internal/trader"
	"cb_grok/pkg/postgres"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"time"
)

// DBMetricsCollector implements metrics.Collector interface and collects metrics directly to PostgreSQL
type DBMetricsCollector struct {
	state       trader.State
	metricsRepo *repository.MetricsRepository
	symbol      string
	logger      *zap.Logger
}

func NewDBMetricsCollector(state trader.State, db postgres.Postgres, symbol string, logger *zap.Logger) *DBMetricsCollector {
	return &DBMetricsCollector{
		state:       state,
		metricsRepo: repository.NewMetricsRepository(db),
		symbol:      symbol,
		logger:      logger,
	}
}

func (m *DBMetricsCollector) SaveIndicatorData(timestamp int64, indicators map[string]float64) error {
	// Save all indicators to time_series_metrics table
	timestampTime := time.UnixMilli(timestamp)

	for name, value := range indicators {
		if err := m.metricsRepo.SaveTimeSeriesMetric(
			timestampTime,
			m.symbol,
			"indicator_"+name,
			value,
			map[string]interface{}{
				"indicator":        name,
				"candle_timestamp": timestamp,
			},
		); err != nil {
			m.logger.Error("Failed to save indicator",
				zap.String("indicator", name),
				zap.Float64("value", value),
				zap.Error(err))
			continue // Continue with other indicators even if one fails
		}
	}

	return nil
}

func (m *DBMetricsCollector) SaveTradeMetric(order trader.Action, indicators map[string]float64) error {
	// Get current performance metrics
	winRate := m.state.CalculateWinRate()
	maxDrawdown := m.state.CalculateMaxDrawdown()
	sharpeRatio := m.state.CalculateSharpeRatio()

	indicatorsJSON, err := json.Marshal(indicators)
	if err != nil {
		return err
	}

	portfolioValue := m.state.GetPortfolioValue()
	profit := order.Profit

	decisionTrigger := string(order.DecisionTrigger)
	metric := &metrics.TradeMetric{
		Timestamp:       time.UnixMilli(order.Timestamp),
		Symbol:          m.symbol,
		Side:            string(order.Decision),
		Price:           order.Price,
		Quantity:        order.AssetAmount,
		Profit:          &profit,
		PortfolioValue:  &portfolioValue,
		Indicators:      indicatorsJSON,
		DecisionTrigger: &decisionTrigger,
		WinRate:         &winRate,
		MaxDrawdown:     &maxDrawdown,
		SharpeRatio:     &sharpeRatio,
	}

	return m.metricsRepo.SaveTradeMetric(context.Background(), metric)
}

func (m *DBMetricsCollector) Close() error {
	m.logger.Info("Metrics collector closed")
	return nil
}
