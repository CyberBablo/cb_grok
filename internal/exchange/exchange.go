package exchange

import (
	"cb_grok/pkg/models"
)

type Exchange interface {
	FetchSpotOHLCV(symbol string, timeframe Timeframe, total int) ([]models.OHLCV, error)
	PlaceSpotMarketOrder(symbol string, orderSide OrderSide, orderAmount float64) error
}
