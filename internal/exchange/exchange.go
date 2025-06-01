package exchange

import (
	"cb_grok/pkg/models"
	"context"
	"github.com/govalues/decimal"
)

type Exchange interface {
	FetchSpotOHLCV(symbol string, timeframe Timeframe, total int) ([]models.OHLCV, error)
	PlaceSpotMarketOrder(ctx context.Context, symbol string, orderSide OrderSide, amount decimal.Decimal) error
}
