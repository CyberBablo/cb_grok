package exchange

import (
	"cb_grok/pkg/models"
	"context"
	"github.com/govalues/decimal"
)

type mock struct {
}

func NewMockExchange() Exchange {
	return &mock{}
}

func (e *mock) FetchSpotOHLCV(symbol string, timeframe Timeframe, limit int) ([]models.OHLCV, error) {
	return nil, nil
}

func (e *mock) PlaceSpotMarketOrder(ctx context.Context, symbol string, orderSide OrderSide, amount decimal.Decimal) error {

	return nil
}

func (e *mock) FetchBalance() (map[string]float64, error) {
	return map[string]float64{}, nil
}
