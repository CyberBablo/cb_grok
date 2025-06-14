package order

import (
	"time"
)

// Order represents the order table
type Order struct {
	ID         int64      `db:"id"`
	ExchangeID int        `db:"exch_id"`
	ProductID  int        `db:"prod_id"`
	TypeID     int        `db:"type_id"`
	SideID     int        `db:"side_id"`
	StatusID   int        `db:"status_id"`
	BaseQty    *float64   `db:"base_qty"`
	QuoteQty   *float64   `db:"quote_qty"`
	ExtID      string     `db:"ext_id"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  *time.Time `db:"updated_at"`
}

// Exchange represents the exchange table
type Exchange struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

const (
	OrderStatusNew      int = 1
	OrderStatusPlaced   int = 2
	OrderStatusFilled   int = 3
	OrderStatusCanceled int = 4
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
