package usecase

import (
	"context"
	"errors"
)

func (u *usecase) UpdateOrderStatus(ctx context.Context, orderID int64, statusID int32) error {
	if u.ex == nil {
		return errors.New("exchange not set")
	}

	// Update order status in database
	if err := u.repo.UpdateOrderStatus(orderID, statusID); err != nil {
		return err
	}

	// TODO: Implement exchange API call if needed
	// exchangeClient.UpdateOrderStatus(orderID, statusID)

	return nil
}
