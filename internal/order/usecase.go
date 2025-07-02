package order

import (
	"cb_grok/internal/exchange"
	order_model "cb_grok/internal/models/order"
	"context"
)

// Usecase interface definition
type Usecase interface {
	Init(ex exchange.Exchange)

	CreateSpotMarketOrder(symbol string, side exchange.OrderSide, baseQty float64, takeProfit *float64, stopLoss *float64) error
	SyncOrders(ctx context.Context)
	GetActiveOrders(ctx context.Context) ([]order_model.Order, error)
	GetSymbolByCode(code string) (*order_model.Symbol, error)
	GetLastOrder() (*order_model.Order, error)
}
