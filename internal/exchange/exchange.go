package exchange

import (
	"cb_grok/pkg/models"
)

type Exchange interface {
	Name() string
	FetchSpotOHLCV(symbol string, timeframe Timeframe, total int) ([]models.OHLCV, error)
	PlaceSpotMarketOrder(symbol string, orderSide OrderSide, quoteQty float64, takeProfit *float64, stopLoss *float64) (string, error)
}
