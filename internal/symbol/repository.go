package symbol

import symbol_model "cb_grok/internal/symbol/model"

type Repository interface {
	GetSymbolByID(id int64) (*symbol_model.Symbol, error)
}
