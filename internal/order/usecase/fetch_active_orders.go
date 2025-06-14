package usecase

import (
	"cb_grok/internal/order"
	"context"
)

func (u *usecase) GetActiveOrders(ctx context.Context) ([]order.Order, error) {
	orders, err := u.repo.GetActiveOrders()
	if err != nil {
		return nil, err
	}
	return orders, nil
}
