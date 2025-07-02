package exchange

import (
	candle_model "cb_grok/internal/models/candle"
	order_model "cb_grok/internal/models/order"
)

type Exchange interface {
	Name() string
	FetchSpotOHLCV(symbol string, timeframe Timeframe, total int) ([]candle_model.OHLCV, error)
	PlaceSpotMarketOrder(symbol string, orderSide OrderSide, baseQty float64, takeProfit *float64, stopLoss *float64) (string, error)
	GetOrderStatus(orderId string) (order_model.OrderStatus, error)
	GetOrderQuoteQty(orderId string) (float64, error)
	GetAvailableSpotWalletBalance(coin string) (float64, error)
}
