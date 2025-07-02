package repository

import (
	"cb_grok/internal/candle"
	candle_model "cb_grok/internal/models/candle"
	"cb_grok/pkg/postgres"
	"context"
	"fmt"
)

type repo struct {
	db postgres.Postgres
}

func New(db postgres.Postgres) candle.Repository {
	return &repo{db: db}
}

func (r *repo) Create(ctx context.Context, symbol, exchange, timeframe string, candle candle_model.OHLCV) error {
	query := `
		INSERT INTO candles (symbol, exchange, timeframe, timestamp, open, high, low, close, volume)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (symbol, exchange, timeframe, timestamp) 
		DO UPDATE SET 
			open = EXCLUDED.open,
			high = EXCLUDED.high,
			low = EXCLUDED.low,
			close = EXCLUDED.close,
			volume = EXCLUDED.volume,
			updated_at = CURRENT_TIMESTAMP
		;
	`

	_, err := r.db.Exec(query, symbol, exchange, timeframe, candle.Timestamp,
		candle.Open, candle.High, candle.Low, candle.Close, candle.Volume)
	if err != nil {
		return fmt.Errorf("failed to save candle: %w", err)
	}

	return nil
}

func (r *repo) Select(ctx context.Context, symbol, exchange, timeframe string, startTime, endTime int64) ([]candle_model.OHLCV, error) {
	query := `
		SELECT
			timestamp,
			open,
			high,
			low,
			close,
			volume
		FROM candles
		WHERE TRUE
			AND symbol = $1
			AND exchange = $2
			AND timeframe = $3 
			AND timestamp >= $4
			AND timestamp <= $5
		ORDER BY timestamp ASC
		;
	`

	rows, err := r.db.Query(query, symbol, exchange, timeframe, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to query candles: %w", err)
	}
	defer rows.Close()

	var result []candle_model.OHLCV
	for rows.Next() {
		var c candle_model.OHLCV
		err := rows.Scan(&c.Timestamp, &c.Open, &c.High, &c.Low, &c.Close, &c.Volume)
		if err != nil {
			return nil, fmt.Errorf("failed to scan candle: %w", err)
		}
		result = append(result, c)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return result, nil
}
