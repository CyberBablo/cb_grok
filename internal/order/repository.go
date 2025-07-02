package order

import order_model "cb_grok/internal/models/order"

// Repository interface
type Repository interface {
	InsertOrder(order *order_model.Order) error
	UpdateOrderStatus(orderID int64, statusID int) error
	UpdateOrderQuoteQty(orderID int64, quoteQty float64) error
	GetActiveOrders() ([]order_model.Order, error)
	GetLastOrder() (*order_model.Order, error)
	GetExchangeByName(name string) (*order_model.Exchange, error)
	UpdateOrderExtID(orderID int64, extID string) error
	GetSymbolByCode(code string) (*order_model.Symbol, error)
}
