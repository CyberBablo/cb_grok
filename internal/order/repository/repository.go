package repository

import (
	"cb_grok/internal/order"
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

func (r *repo) InsertOrder(order *order.Order) error {
	query := `
		INSERT INTO public.order (
			exch_id, prod_id, type_id, side_id, status_id, 
			base_qty, quote_qty, ext_id, created_at, updated_at, tp_price, sl_price
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`
	var id int64
	err := r.db.Get(&id, query,
		order.ExchangeID,
		order.ProductID,
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

func (r *repo) GetActiveOrders() ([]order.Order, error) {
	var orders []order.Order
	query := `
		SELECT o.id, o.exch_id, o.prod_id, o.type_id, o.side_id, o.status_id,
			o.base_qty, o.quote_qty, o.ext_id, o.created_at, o.updated_at, tp_price, sl_price
		FROM public.order o
		JOIN public.order_status os ON o.status_id = os.id
		WHERE os.code IN ('new', 'placed')
	`
	err := r.db.Select(&orders, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active orders: %w", err)
	}
	return orders, nil
}

func (r *repo) GetExchangeByName(name string) (*order.Exchange, error) {
	var result []order.Exchange
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
