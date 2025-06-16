package usecase

import (
	"context"
	"go.uber.org/zap"
	"time"
)

func (u *usecase) SyncOrders(ctx context.Context) {
	if u.ex == nil {
		return
	}

	ticker := time.NewTicker(5 * time.Second)
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
				exchangeStatus, err := u.ex.GetOrderInfo(order.ExtID)
				if err != nil {
					u.log.Error("failed to get order info", zap.String("order_id", order.ExtID), zap.Error(err))
					continue
				}
				if int(exchangeStatus) != order.StatusID {
					err := u.repo.UpdateOrderStatus(order.ID, int(exchangeStatus))
					if err != nil {
						u.log.Error("failed to update order status", zap.String("order_id", order.ExtID), zap.Error(err))
						continue
					}
				}
			}
		}
	}
}
