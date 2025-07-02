package candle

import "time"

type CandleModel struct {
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
