package usecase

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/order"
	"go.uber.org/zap"
)

type usecase struct {
	repo order.Repository
	ex   exchange.Exchange
	log  *zap.Logger
}

func New(repo order.Repository, log *zap.Logger) order.Usecase {
	return &usecase{
		repo: repo,
		log:  log,
	}
}

func (u *usecase) SetExchange(ex exchange.Exchange) {
	u.ex = ex
}
