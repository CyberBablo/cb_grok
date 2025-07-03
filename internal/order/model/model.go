package order_model

import (
	"time"
)

// Order represents the order table
type Order struct {
	ID              int64      `db:"id"`
	SymbolID        int64      `db:"symbol_id"`
	ExchangeID      int64      `db:"exch_id"`
	TypeID          int64      `db:"type_id"`
	SideID          int64      `db:"side_id"`
	StatusID        int64      `db:"status_id"`
	BaseQty         *float64   `db:"base_qty"`
	QuoteQty        *float64   `db:"quote_qty"`
	ExtID           string     `db:"ext_id"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       *time.Time `db:"updated_at"`
	TakeProfitPrice *float64   `db:"tp_price"`
	StopLossPrice   *float64   `db:"sl_price"`
	TraderID        int64      `db:"trader_id"`
}

// Exchange represents the exchange table
type Exchange struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
}

type Symbol struct {
	ID    int64  `db:"id"`
	Code  string `db:"code"`
	Base  string `db:"base"`
	Quote string `db:"quote"`
}

type OrderStatus int64

const (
	OrderStatusNew      OrderStatus = 1
	OrderStatusPlaced   OrderStatus = 2
	OrderStatusFilled   OrderStatus = 3
	OrderStatusCanceled OrderStatus = 4
)

const (
	OrderProductSpot int64 = 1
)

const (
	OrderTypeMarket int64 = 1
)

type OrderSide int64

const (
	OrderSideBuy  OrderSide = 1
	OrderSideSell OrderSide = 2
)
