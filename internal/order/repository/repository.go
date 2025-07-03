package repository

import (
	"cb_grok/internal/order"
	order_model "cb_grok/internal/order/model"
	"cb_grok/pkg/postgres"
	"errors"
	"fmt"
)

type repo struct {
	db postgres.Postgres
}

func New(db postgres.Postgres) order.Repository {
	return &repo{
		db: db,
	}
}

func (r *repo) InsertOrder(order *order_model.Order) error {
	query := `
		INSERT INTO public.order (
			symbol_id, exch_id, type_id, side_id, status_id, 
			base_qty, quote_qty, ext_id, created_at, updated_at, tp_price, sl_price, trader_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`
	var id int64
	err := r.db.Get(&id, query,
		order.SymbolID,
		order.ExchangeID,
		order.TypeID,
		order.SideID,
		order.StatusID,
		order.BaseQty,
		order.QuoteQty,
		order.ExtID,
		order.CreatedAt,
		order.UpdatedAt,
		order.TakeProfitPrice,
		order.StopLossPrice,
		order.TraderID,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}
	order.ID = id
	return nil
}

func (r *repo) UpdateOrderStatus(orderID int64, statusID int) error {
	query := `
		UPDATE public.order 
		SET status_id = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id
	`
	var id int64
	err := r.db.Get(&id, query, statusID, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	if id == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *repo) UpdateOrderQuoteQty(orderID int64, quoteQty float64) error {
	query := `
		UPDATE public.order 
		SET quote_qty = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING id
	`
	var id int64
	err := r.db.Get(&id, query, quoteQty, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}
	if id == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *repo) GetActiveOrders() ([]order_model.Order, error) {
	var orders []order_model.Order
	query := `
		SELECT o.id, o.symbol_id, o.exch_id, o.type_id, o.side_id, o.status_id,
			o.base_qty, o.quote_qty, o.ext_id, o.created_at, o.updated_at, o.tp_price, o.sl_price, o.trader_id
		FROM public.order o
		JOIN public.order_status os ON o.status_id = os.id
		WHERE os.code IN ('new', 'placed', 'filled')
	`
	err := r.db.Select(&orders, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}
	return orders, nil
}

func (r *repo) GetLastOrder(traderID int64) (*order_model.Order, error) {
	var orders []order_model.Order
	query := `
		SELECT o.id, o.symbol_id, o.exch_id, o.type_id, o.side_id, o.status_id,
			o.base_qty, o.quote_qty, o.ext_id, o.created_at, o.updated_at, o.tp_price, o.sl_price
		FROM public.order o where trader_id=$1 ORDER BY o.id desc LIMIT 1
	`
	err := r.db.Select(&orders, query, traderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}
	if len(orders) == 0 {
		return nil, nil
	}

	return &orders[0], nil
}

func (r *repo) GetExchangeByName(name string) (*order_model.Exchange, error) {
	var result []order_model.Exchange
	query := `
		SELECT id, name 
		FROM public.exchange 
		WHERE name = $1
	`
	err := r.db.Select(&result, query, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange by name: %w", err)
	}
	if len(result) == 0 {
		return nil, errors.New("exchange not found")
	}
	return &result[0], nil
}

func (r *repo) UpdateOrderExtID(orderID int64, extID string) error {
	query := `
		UPDATE public.order 
		SET ext_id = $1
		WHERE id = $2
		RETURNING id
	`
	var id int64
	err := r.db.Get(&id, query, extID, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order ext_id: %w", err)
	}
	if id == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *repo) GetSymbolByCode(code string) (*order_model.Symbol, error) {
	var result []order_model.Symbol
	query := `
		SELECT id, code, base, quote 
		FROM public.symbol 
		WHERE code = $1
	`
	err := r.db.Select(&result, query, code)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol by code: %w", err)
	}
	if len(result) == 0 {
		return nil, errors.New("symbol not found")
	}
	return &result[0], nil
}
