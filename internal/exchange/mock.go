package exchange

import (
	"cb_grok/internal/order/model"
	"cb_grok/pkg/models"
)

type mock struct {
}

func NewMockExchange() Exchange {
	return &mock{}
}
func (m *mock) Name() string {
	return "mock"
}

func (m *mock) FetchSpotOHLCV(symbol string, timeframe Timeframe, total int) ([]models.OHLCV, error) {
	return []models.OHLCV{
		{
			Timestamp: 0,
			Open:      0,
			High:      0,
			Low:       0,
			Close:     0,
			Volume:    0,
		},
	}, nil
}
func (m *mock) PlaceSpotMarketOrder(symbol string, orderSide OrderSide, baseQty float64, takeProfit *float64, stopLoss *float64) (string, error) {
	return "mock-order-id", nil
}
func (m *mock) GetOrderInfo(orderId string) (order_model.OrderStatus, error) {
	return order_model.OrderStatusFilled, nil
}
func (m *mock) GetAvailableSpotWalletBalance(coin string) (float64, error) {
	return 1000.0, nil
}
