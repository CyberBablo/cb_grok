package usecase

import (
	order_model "cb_grok/internal/order/model"
	"context"
)

func (u *orderUC) GetActiveOrders(ctx context.Context) ([]order_model.Order, error) {
	orders, err := u.repo.GetActiveOrders()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (u *orderUC) GetLastOrder(traderID int64) (*order_model.Order, error) {
	return u.repo.GetLastOrder(traderID)
}
