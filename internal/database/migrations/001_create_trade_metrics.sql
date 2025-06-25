-- Create trade_metrics table for storing detailed trade information
CREATE TABLE IF NOT EXISTS trade_metrics (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    symbol VARCHAR(20) NOT NULL,
    side VARCHAR(10) NOT NULL, -- 'BUY' or 'SELL'
    price DECIMAL(20, 8) NOT NULL,
    quantity DECIMAL(20, 8) NOT NULL,
    profit DECIMAL(20, 8),
    portfolio_value DECIMAL(20, 8),

    -- Strategy parameters at time of trade
    strategy_params JSONB,

    -- Indicator values at time of trade
    indicators JSONB,

    -- Decision metadata
    decision_trigger VARCHAR(100),
    signal_strength DECIMAL(5, 2),

    -- Performance metrics
    win_rate DECIMAL(5, 2),
    max_drawdown DECIMAL(5, 2),
    sharpe_ratio DECIMAL(10, 4),

    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_trade_metrics__time__symbol ON trade_metrics(timestamp, symbol);
CREATE INDEX idx_trade_metrics__time__side ON trade_metrics(timestamp, side);
CREATE INDEX idx_trade_metrics_created_at ON trade_metrics(created_at);

-- Create strategy_runs table to track different strategy executions
CREATE TABLE IF NOT EXISTS strategy_runs (
    id BIGSERIAL PRIMARY KEY,
    run_id UUID DEFAULT gen_random_uuid(),
    symbol VARCHAR(20) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    initial_capital DECIMAL(20, 8) NOT NULL,
    final_capital DECIMAL(20, 8),

    -- Strategy configuration
    strategy_type VARCHAR(50) NOT NULL,
    strategy_params JSONB NOT NULL,

    -- Performance summary
    total_trades INTEGER DEFAULT 0,
    winning_trades INTEGER DEFAULT 0,
    losing_trades INTEGER DEFAULT 0,
    total_profit DECIMAL(20, 8) DEFAULT 0,
    max_drawdown DECIMAL(5, 2),
    sharpe_ratio DECIMAL(10, 4),
    win_rate DECIMAL(5, 2),

    -- Metadata
    environment VARCHAR(20) DEFAULT 'backtest', -- 'backtest', 'paper', 'live'
    notes TEXT,

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_strategy_runs_symbol ON strategy_runs(symbol);
CREATE INDEX idx_strategy_runs_run_id ON strategy_runs(run_id);
CREATE INDEX idx_strategy_runs_created_at ON strategy_runs(created_at);

-- Create time_series_metrics table for continuous metrics
CREATE TABLE IF NOT EXISTS time_series_metrics (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    symbol VARCHAR(20) NOT NULL,
    metric_name VARCHAR(50) NOT NULL,
    metric_value DECIMAL(20, 8) NOT NULL,
    labels JSONB,

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_time_series_metrics__main ON time_series_metrics(timestamp, symbol, metric_name);
CREATE INDEX idx_time_series_metrics__created_at ON time_series_metrics(created_at);
