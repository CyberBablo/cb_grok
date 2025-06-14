package order

import (
	"cb_grok/internal/exchange"
	"context"
)

// Usecase interface definition
type Usecase interface {
	SetExchange(ex exchange.Exchange)

	CreateSpotMarketOrder(symbol string, side exchange.OrderSide, quoteQty float64) error
	UpdateOrderStatus(ctx context.Context, orderID int64, statusID int32) error
	SyncOrders(ctx context.Context)
	GetActiveOrders(ctx context.Context) ([]Order, error)
}
