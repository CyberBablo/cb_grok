package repository

import (
	"cb_grok/internal/symbol"
	symbol_model "cb_grok/internal/symbol/model"
	"cb_grok/pkg/postgres"
	"errors"
	"fmt"
)

type repo struct {
	db postgres.Postgres
}

func (r repo) GetSymbolByID(id int64) (*symbol_model.Symbol, error) {
	var result []*symbol_model.Symbol
	query := `
		SELECT id, code, prod_id, base, quote, decimals
		FROM public.symbol 
		WHERE id = $1
	`
	err := r.db.Select(&result, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get symbol by id: %d %w ", id, err)
	}
	if len(result) == 0 {
		return nil, errors.New("symbol not found")
	}
	return result[0], nil
}

func New(db postgres.Postgres) symbol.Repository {
	return &repo{
		db: db,
	}
}
