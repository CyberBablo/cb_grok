package repository

import (
	"cb_grok/internal/metrics"
	"cb_grok/pkg/postgres"
	"context"
	"encoding/json"
	"time"
)

// MetricsRepository implements metrics.Repository interface
type MetricsRepository struct {
	db postgres.Postgres
}

func NewMetricsRepository(db postgres.Postgres) *MetricsRepository {
	return &MetricsRepository{db: db}
}

func (r *MetricsRepository) SaveTradeMetric(ctx context.Context, metric *metrics.TradeMetric) error {
	query := `
		INSERT INTO trade_metrics (
			timestamp, symbol, side, price, quantity, profit, portfolio_value,
			strategy_params, indicators, decision_trigger, signal_strength,
			win_rate, max_drawdown, sharpe_ratio
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	// Handle nil values
	if metric.StrategyParams == nil {
		metric.StrategyParams = json.RawMessage("{}")
	}
	if metric.Indicators == nil {
		metric.Indicators = json.RawMessage("{}")
	}

	_, err := r.db.Exec(query,
		metric.Timestamp, metric.Symbol, metric.Side, metric.Price, metric.Quantity,
		metric.Profit, metric.PortfolioValue, metric.StrategyParams, metric.Indicators,
		metric.DecisionTrigger, metric.SignalStrength, metric.WinRate,
		metric.MaxDrawdown, metric.SharpeRatio,
	)
	return err
}

func (r *MetricsRepository) CreateStrategyRun(ctx context.Context, run *metrics.StrategyRun) error {
	query := `
		INSERT INTO strategy_runs (
			run_id, symbol, start_time, initial_capital, strategy_type, strategy_params, environment
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
	`
	_, err := r.db.Exec(query,
		run.RunID, run.Symbol, run.StartTime, run.InitialCapital,
		run.StrategyType, run.StrategyParams, run.Environment,
	)
	return err
}

func (r *MetricsRepository) UpdateStrategyRun(ctx context.Context, run *metrics.StrategyRun) error {
	query := `
		UPDATE strategy_runs SET
			end_time = $2,
			final_capital = $3,
			total_trades = $4,
			winning_trades = $5,
			losing_trades = $6,
			total_profit = $7,
			max_drawdown = $8,
			sharpe_ratio = $9,
			win_rate = $10,
			updated_at = NOW()
		WHERE run_id = $1
	`
	_, err := r.db.Exec(query,
		run.RunID, run.EndTime, run.FinalCapital, run.TotalTrades,
		run.WinningTrades, run.LosingTrades, run.TotalProfit,
		run.MaxDrawdown, run.SharpeRatio, run.WinRate,
	)
	return err
}

func (r *MetricsRepository) SaveTimeSeriesMetric(timestamp time.Time, symbol, metricName string, value float64, labels map[string]interface{}) error {
	labelsJSON, err := json.Marshal(labels)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO time_series_metrics (timestamp, symbol, metric_name, metric_value, labels)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = r.db.Exec(query, timestamp, symbol, metricName, value, labelsJSON)
	return err
}
