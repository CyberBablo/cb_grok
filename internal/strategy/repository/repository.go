package repository

import (
	"cb_grok/internal/strategy"
	strategyModel "cb_grok/internal/strategy/model"
	"cb_grok/pkg/postgres"
	"encoding/json"
	"errors"
	"fmt"
)

type repo struct {
	db postgres.Postgres
}

func New(db postgres.Postgres) strategy.Repository {
	return &repo{
		db: db,
	}
}

func (r *repo) InsertStrategy(entity *strategyModel.Strategy) error {
	query := `
		INSERT INTO public.strategy (
			symbol_id, params
		) VALUES ($1, $2)
		RETURNING id
	`
	var id int64
	paramsString, err := json.Marshal(entity.Params)
	if err != nil {
		return fmt.Errorf("Failed to marshal strategy params%w", err)
	}
	err = r.db.Get(&id, query,
		entity.SymbolID,
		paramsString,
	)
	if err != nil {
		return fmt.Errorf("failed to insert strategy: %w", err)
	}
	return nil
}

func (r *repo) GetStrategy(id int64) (*strategyModel.Strategy, error) {
	var result []strategyModel.Strategy
	query := `
		SELECT id, symbol_id, created_at, params, timeframe
		FROM public.strategy 
		WHERE id = $1
	`
	err := r.db.Select(&result, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get exchange by name: %w", err)
	}
	if len(result) == 0 {
		return nil, errors.New("exchange not found")
	}
	return &result[0], nil
}
