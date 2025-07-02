package usecase

import (
	order_model "cb_grok/internal/models/order"
	"context"
)

func (u *usecase) GetActiveOrders(ctx context.Context) ([]order_model.Order, error) {
	orders, err := u.repo.GetActiveOrders()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (u *usecase) GetLastOrder() (*order_model.Order, error) {
	return u.repo.GetLastOrder()
}
