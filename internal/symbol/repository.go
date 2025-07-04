package symbol

import symbolModel "cb_grok/internal/symbol/model"

type Repository interface {
	GetSymbolByID(id int64) (*symbolModel.Symbol, error)
}
