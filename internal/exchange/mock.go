package exchange

import (
	"cb_grok/pkg/models"
)

type mockImpl struct {
}

func NewMockExchange() Exchange {
	return &mockImpl{}
}

func (e *mockImpl) FetchOHLCV(symbol, timeframe string, limit int) ([]models.OHLCV, error) {
	return nil, nil
}

func (e *mockImpl) CreateOrder(symbol, side string, amount float64, stopLoss, takeProfit float64) error {

	return nil
}

func (e *mockImpl) FetchBalance() (map[string]float64, error) {
	return map[string]float64{}, nil
}

func (e *mockImpl) GetWSUrl() string {
	return "ws://localhost:8080/ws"
}
