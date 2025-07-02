package metrics

import (
	"encoding/json"
	"time"
)

// TradeMetric represents a single trade metric entry
type TradeMetric struct {
	ID              int64           `db:"id"`
	Timestamp       time.Time       `db:"timestamp"`
	Symbol          string          `db:"symbol"`
	Side            string          `db:"side"`
	Price           float64         `db:"price"`
	Quantity        float64         `db:"quantity"`
	Profit          *float64        `db:"profit"`
	PortfolioValue  *float64        `db:"portfolio_value"`
	StrategyParams  json.RawMessage `db:"strategy_params"`
	Indicators      json.RawMessage `db:"indicators"`
	DecisionTrigger *string         `db:"decision_trigger"`
	SignalStrength  *float64        `db:"signal_strength"`
	WinRate         *float64        `db:"win_rate"`
	MaxDrawdown     *float64        `db:"max_drawdown"`
	SharpeRatio     *float64        `db:"sharpe_ratio"`
	CreatedAt       time.Time       `db:"created_at"`
}

// StrategyRun represents a strategy execution run
type StrategyRun struct {
	ID             int64           `db:"id"`
	RunID          string          `db:"run_id"`
	Symbol         string          `db:"symbol"`
	StartTime      time.Time       `db:"start_time"`
	EndTime        *time.Time      `db:"end_time"`
	InitialCapital float64         `db:"initial_capital"`
	FinalCapital   *float64        `db:"final_capital"`
	StrategyType   string          `db:"strategy_type"`
	StrategyParams json.RawMessage `db:"strategy_params"`
	TotalTrades    int             `db:"total_trades"`
	WinningTrades  int             `db:"winning_trades"`
	LosingTrades   int             `db:"losing_trades"`
	TotalProfit    float64         `db:"total_profit"`
	MaxDrawdown    *float64        `db:"max_drawdown"`
	SharpeRatio    *float64        `db:"sharpe_ratio"`
	WinRate        *float64        `db:"win_rate"`
	Environment    string          `db:"environment"`
	Notes          *string         `db:"notes"`
	CreatedAt      time.Time       `db:"created_at"`
	UpdatedAt      time.Time       `db:"updated_at"`
}
