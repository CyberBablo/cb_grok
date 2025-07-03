package model

type Trader struct {
	ID         int64   `db:"id"`
	SymbolID   int64   `db:"symbol_id"`
	InitQty    float64 `db:"init_qty"`
	StrategyID int64   `db:"strategy_id"`
	StageID    int64   `db:"stage_id"`
}
