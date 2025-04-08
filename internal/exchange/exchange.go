package exchange

import (
	"cb_grok/pkg/models"
)

type Exchange interface {
	FetchOHLCV(symbol string, timeframe string, total int) ([]models.OHLCV, error)
	CreateOrder(symbol, side string, amount, stopLoss, takeProfit float64) error

	GetWSUrl() string
}
