package usecase

import (
	"context"
	"time"
)

func (u *usecase) SyncOrders(ctx context.Context) {
	if u.ex == nil {
		return
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get active orders from database
			orders, err := u.repo.GetActiveOrders()
			if err != nil {
				// TODO: Implement proper error logging
				continue
			}

			for _, order := range orders {
				_ = order
				// TODO: Implement exchange API call to check order status
				// exchangeStatus := exchangeClient.GetOrderStatus(order.ID)
				// if exchangeStatus != order.StatusID {
				//     u.repo.UpdateOrderStatus(order.ID, exchangeStatus)
				// }
			}
		}
	}
}
