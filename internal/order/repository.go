package order

// Repository interface
type Repository interface {
	InsertOrder(order *Order) error
	UpdateOrderStatus(orderID int64, statusID int32) error
	GetActiveOrders() ([]Order, error)
	GetExchangeByName(name string) (*Exchange, error)
	UpdateOrderExtID(orderID int64, extID string) error
}
