package usecase

import (
	"cb_grok/internal/order/model"
)

func (u *orderUC) GetSymbolByCode(code string) (*order_model.Symbol, error) {
	return u.repo.GetSymbolByCode(code)
}
