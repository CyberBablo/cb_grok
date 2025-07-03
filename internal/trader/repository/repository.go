package repository

import (
	"cb_grok/internal/trader"
	"cb_grok/internal/trader/model"
	"cb_grok/pkg/postgres"
	"errors"
	"fmt"
)

type repo struct {
	db postgres.Postgres
}

func New(db postgres.Postgres) trader.Repository {
	return &repo{
		db: db,
	}
}

func (r repo) GetTraderByStage(stageID int) ([]*model.Trader, error) {
	var result []*model.Trader
	query := `
		SELECT id, symbol_id, init_qty, strategy_id, stage_id
		FROM public.trader 
		WHERE stage_id = $1
	`
	err := r.db.Select(&result, query, stageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get trader by stage: %w", err)
	}
	if len(result) == 0 {
		return nil, errors.New(fmt.Sprintf("trader not found by stage ID %d", stageID))
	}
	return result, nil
}
