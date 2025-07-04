package usecase

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/order/model"
	symbolModel "cb_grok/internal/symbol/model"
	"errors"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"time"
)

func (u *orderUC) CreateSpotMarketOrder(symbol symbolModel.Symbol, side exchange.OrderSide, baseQty float64, takeProfit *float64, stopLoss *float64, traderID int64) error {
	if u.ex == nil {
		return errors.New("exchange not set")
	}

	exch, err := u.repo.GetExchangeByName(u.ex.Name())
	if err != nil {
		u.log.Error("failed to get exchange by name", zap.Error(err))
		return err
	}

	sideId := int64(1)
	if side == exchange.OrderSideSell {
		sideId = int64(2)
	}

	symbolValue, err := u.repo.GetSymbolByCode(symbol.Code)
	if err != nil {
		u.log.Error("failed to get symbol by code", zap.Error(err))
		return err
	}

	// Insert order into database
	ord := &order_model.Order{
		ExchangeID:      exch.ID,
		SymbolID:        symbolValue.ID,
		TypeID:          order_model.OrderTypeMarket,
		SideID:          sideId,
		StatusID:        int64(order_model.OrderStatusNew),
		BaseQty:         lo.ToPtr(baseQty),
		QuoteQty:        nil,
		ExtID:           "",
		CreatedAt:       time.Now(),
		UpdatedAt:       nil,
		TakeProfitPrice: takeProfit,
		StopLossPrice:   stopLoss,
		TraderID:        traderID,
	}
	err = u.repo.InsertOrder(ord)
	if err != nil {
		return err
	}

	orderId, err := u.ex.PlaceSpotMarketOrder(symbol.Code, side, baseQty, nil, nil, symbol.Decimals)
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
