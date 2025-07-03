package usecase

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/order"
	"context"
	"go.uber.org/zap"
)

type orderUC struct {
	repo order.Repository
	ex   exchange.Exchange
	log  *zap.Logger
}

func New(repo order.Repository, log *zap.Logger) order.Order {
	return &orderUC{
		repo: repo,
		log:  log,
	}
}

func (u *orderUC) Init(ex exchange.Exchange) {
	u.ex = ex

	go u.SyncOrders(context.Background())
}
