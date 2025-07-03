package strategy

import strategyModel "cb_grok/internal/strategy/model"

type Repository interface {
	InsertStrategy(entity *strategyModel.Strategy) error
	GetStrategy(id int64) (*strategyModel.Strategy, error)
}
