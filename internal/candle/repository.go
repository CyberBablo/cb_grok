package candle

import (
	candle_model "cb_grok/internal/models/candle"
	"context"
)

type Repository interface {
	Create(ctx context.Context, symbol, exchange, timeframe string, candle candle_model.OHLCV) error
	Select(ctx context.Context, symbol, exchange, timeframe string, startTime, endTime int64) ([]candle_model.OHLCV, error)
}
