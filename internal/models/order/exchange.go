package order

type OrderStatus int

const (
	OrderStatusNew      OrderStatus = 1
	OrderStatusPlaced   OrderStatus = 2
	OrderStatusFilled   OrderStatus = 3
	OrderStatusCanceled OrderStatus = 4
)

const (
	OrderProductSpot int = 1
)

const (
	OrderTypeMarket int = 1
)

type OrderSide int

const (
	OrderSideBuy  OrderSide = 1
	OrderSideSell OrderSide = 2
)
