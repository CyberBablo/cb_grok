package usecase

import (
	order_model "cb_grok/internal/order/model"
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"
)

func (u *orderUC) SyncOrders(ctx context.Context) {
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
				exchangeStatus, err := u.ex.GetOrderStatus(order.ExtID)
				if err != nil {
					u.log.Error("failed to get order info", zap.String("order_id", order.ExtID), zap.Error(err))
					continue
				}
				if int64(exchangeStatus) != order.StatusID {
					u.log.Info(fmt.Sprintf("STARTED ORDER SYNCING %s", order.ExtID))
					err := u.repo.UpdateOrderStatus(order.ID, int(exchangeStatus))
					if err != nil {
						u.log.Error("failed to update order status", zap.String("order_id", order.ExtID), zap.Error(err))
						continue
					}
					if int64(exchangeStatus) == int64(order_model.OrderStatusFilled) {
						u.log.Info(fmt.Sprintf("ORDER FILLED %s", order.ExtID))
						quoteQty, err := u.ex.GetOrderQuoteQty(order.ExtID)
						if err != nil {
							u.log.Error("failed to update order quoteQty", zap.String("order_id", order.ExtID), zap.Error(err))
							continue
						}
						err = u.repo.UpdateOrderQuoteQty(order.ID, quoteQty)
						if err != nil {
							u.log.Error("failed to update order quoteQty", zap.String("order_id", order.ExtID), zap.Error(err))
							continue
						}
					}

				}
				if int64(exchangeStatus) == int64(order_model.OrderStatusFilled) && order.QuoteQty == nil {
					u.log.Info(fmt.Sprintf("ORDER MISSED FILLED %s", order.ExtID))
					quoteQty, err := u.ex.GetOrderQuoteQty(order.ExtID)
					if err != nil {
						u.log.Error("failed to update order quoteQty", zap.String("order_id", order.ExtID), zap.Error(err))
						continue
					}
					err = u.repo.UpdateOrderQuoteQty(order.ID, quoteQty)
					if err != nil {
						u.log.Error("failed to update order quoteQty", zap.String("order_id", order.ExtID), zap.Error(err))
						continue
					}
				}
			}
		}
	}
}
