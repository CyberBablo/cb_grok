-- Create candles table for storing OHLCV data
CREATE TABLE IF NOT EXISTS candles (
    id SERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    exchange VARCHAR(50) NOT NULL,
    timeframe VARCHAR(10) NOT NULL,
    timestamp BIGINT NOT NULL,
    open DOUBLE PRECISION NOT NULL,
    high DOUBLE PRECISION NOT NULL,
    low DOUBLE PRECISION NOT NULL,
    close DOUBLE PRECISION NOT NULL,
    volume DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(symbol, exchange, timeframe, timestamp)
);

-- Create indexes for common queries
CREATE INDEX idx_candles_symbol_exchange_timeframe_timestamp 
    ON candles(symbol, exchange, timeframe, timestamp);

CREATE INDEX idx_candles_timestamp 
    ON candles(timestamp);

-- Create updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_candles_updated_at 
    BEFORE UPDATE ON candles 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
