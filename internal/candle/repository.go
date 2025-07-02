package candle

import (
	"cb_grok/pkg/models"
	"context"
)

type Repository interface {
	Create(ctx context.Context, symbol, exchange, timeframe string, candle models.OHLCV) error
	Select(ctx context.Context, symbol, exchange, timeframe string, startTime, endTime int64) ([]models.OHLCV, error)
}
