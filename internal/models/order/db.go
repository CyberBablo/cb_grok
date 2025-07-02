package order

import (
	"time"
)

// Order represents the order table
type Order struct {
	ID              int64      `db:"id"`
	SymbolID        int        `db:"symbol_id"`
	ExchangeID      int        `db:"exch_id"`
	TypeID          int        `db:"type_id"`
	SideID          int        `db:"side_id"`
	StatusID        int        `db:"status_id"`
	BaseQty         *float64   `db:"base_qty"`
	QuoteQty        *float64   `db:"quote_qty"`
	ExtID           string     `db:"ext_id"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       *time.Time `db:"updated_at"`
	TakeProfitPrice *float64   `db:"tp_price"`
	StopLossPrice   *float64   `db:"sl_price"`
}

// Exchange represents the exchange table
type Exchange struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Symbol struct {
	ID    int    `db:"id"`
	Code  string `db:"code"`
	Base  string `db:"base"`
	Quote string `db:"quote"`
}
