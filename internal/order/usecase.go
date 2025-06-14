package order

import (
	"cb_grok/internal/exchange"
	"context"
)

// Usecase interface definition
type Usecase interface {
	Init(ex exchange.Exchange)

	CreateSpotMarketOrder(symbol string, side exchange.OrderSide, quoteQty float64, takeProfit *float64, stopLoss *float64) error
	SyncOrders(ctx context.Context)
	GetActiveOrders(ctx context.Context) ([]Order, error)
}
