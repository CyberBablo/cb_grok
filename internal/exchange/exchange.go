package exchange

import (
	"cb_grok/internal/order/model"
	"cb_grok/pkg/models"
)

type Exchange interface {
	Name() string
	FetchSpotOHLCV(symbol string, timeframe Timeframe, total int) ([]models.OHLCV, error)
	PlaceSpotMarketOrder(symbol string, orderSide OrderSide, baseQty float64, takeProfit *float64, stopLoss *float64) (string, error)
	GetOrderInfo(orderId string) (order_model.OrderStatus, error)
	GetAvailableSpotWalletBalance(coin string) (float64, error)
}
