package usecase

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/order"
	"errors"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"time"
)

func (u *usecase) CreateSpotMarketOrder(symbol string, side exchange.OrderSide, quoteQty float64) error {
	if u.ex == nil {
		return errors.New("exchange not set")
	}

	exch, err := u.repo.GetExchangeByName(u.ex.Name())
	if err != nil {
		u.log.Error("failed to get exchange by name", zap.Error(err))
		return err
	}

	sideId := 1
	if side == exchange.OrderSideSell {
		sideId = 2
	}

	// Insert order into database
	ord := &order.Order{
		ExchangeID: exch.ID,
		ProductID:  order.OrderProductSpot,
		TypeID:     order.OrderTypeMarket,
		SideID:     sideId,
		StatusID:   order.OrderStatusNew,
		BaseQty:    nil,
		QuoteQty:   lo.ToPtr(quoteQty),
		ExtID:      "",
		CreatedAt:  time.Now(),
		UpdatedAt:  nil,
	}
	err = u.repo.InsertOrder(ord)
	if err != nil {
		return err
	}

	orderId, err := u.ex.PlaceSpotMarketOrder(symbol, side, quoteQty)
	if err != nil {
		u.log.Error("create order failed", zap.Error(err))
		return err
	}
	if orderId == "" {
		u.log.Error("create order failed: order ID is empty")
		return errors.New("order ID is empty")
	}

	err = u.repo.UpdateOrderExtID(ord.ID, orderId)
	if err != nil {
		u.log.Error("failed to update order external ID", zap.Int64("order_id", ord.ID), zap.Error(err))
		return err
	}

	return nil
}
