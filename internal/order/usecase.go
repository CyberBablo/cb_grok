package order

import (
	"cb_grok/internal/exchange"
	order_model "cb_grok/internal/order/model"
	"context"
)

// Order interface definition
type Order interface {
	Init(ex exchange.Exchange)

	CreateSpotMarketOrder(symbol string, side exchange.OrderSide, baseQty float64, takeProfit *float64, stopLoss *float64, traderID int64) error
	SyncOrders(ctx context.Context)
	GetActiveOrders(ctx context.Context) ([]order_model.Order, error)
	GetSymbolByCode(code string) (*order_model.Symbol, error)
	GetLastOrder(traderID int64) (*order_model.Order, error)
}
