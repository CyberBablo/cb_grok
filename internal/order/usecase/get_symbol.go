package usecase

import (
	order_model "cb_grok/internal/models/order"
)

func (u *usecase) GetSymbolByCode(code string) (*order_model.Symbol, error) {
	return u.repo.GetSymbolByCode(code)
}
