package repository

import (
	"context"
	"fmt"
	"time"

	"cb_grok/internal/candle"
	"cb_grok/pkg/models"
	"cb_grok/pkg/postgres"
)

type Candle struct {
	ID        int       `db:"id"`
	Symbol    string    `db:"symbol"`
	Exchange  string    `db:"exchange"`
	Timeframe string    `db:"timeframe"`
	Timestamp int64     `db:"timestamp"`
	Open      float64   `db:"open"`
	High      float64   `db:"high"`
	Low       float64   `db:"low"`
	Close     float64   `db:"close"`
	Volume    float64   `db:"volume"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type repository struct {
	db postgres.Postgres
}

func New(db postgres.Postgres) candle.Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, symbol, exchange, timeframe string, candle models.OHLCV) error {
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

func (r *repository) Select(ctx context.Context, symbol, exchange, timeframe string, startTime, endTime int64) ([]models.OHLCV, error) {
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

	var result []models.OHLCV
	for rows.Next() {
		var c models.OHLCV
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
