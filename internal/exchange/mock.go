package exchange

import (
	"cb_grok/pkg/models"
)

type mock struct {
}

func NewMockExchange() Exchange {
	return &mock{}
}

func (e *mock) FetchSpotOHLCV(symbol string, timeframe Timeframe, limit int) ([]models.OHLCV, error) {
	return nil, nil
}

func (e *mock) PlaceSpotMarketOrder(symbol string, orderSide OrderSide, orderAmount float64) error {
	return nil
}

func (e *mock) FetchBalance() (map[string]float64, error) {
	return map[string]float64{}, nil
}
