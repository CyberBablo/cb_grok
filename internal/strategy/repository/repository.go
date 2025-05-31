package repository

import (
	"context"
	"time"

	"cb_grok/internal/strategy"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// StrategyRepository provides methods to interact with strategy data in the database
type StrategyRepository interface {
	// Strategy table methods
	Fetch(
		ctx context.Context,
		id *int64,
		symbol *string,
		timeframe *string,
		dt_left *time.Time,
		dt_right *time.Time,
	) ([]Strategy, error)
	Create(ctx context.Context, strategy StrategyCreate) (int64, error)
	Update(ctx context.Context, id int64, data strategy.StrategyParams) error
	Delete(ctx context.Context, id int64) error

	// StrategyAcc table methods
	FetchStrategyAcc(
		ctx context.Context,
		id *int64,
		strategyID *int64,
		symbol *string,
		dt_upd_left *time.Time,
		dt_upd_right *time.Time,
		dt_create_left *time.Time,
		dt_create_right *time.Time,
	) ([]StrategyAcc, error)
	CreateStrategyAcc(ctx context.Context, strategyID int64) (int64, error)
	UpdateStrategyAcc(ctx context.Context, id int64, strategyID int64) error
	DeleteStrategyAcc(ctx context.Context, id int64) error
}

// StrategyCreate contains all required fields to create a new strategy
type StrategyCreate struct {
	Symbol      string
	Timeframe   string
	Trials      int
	Workers     int
	ValSetDays  int
	TrainSetDay int
	WinRate     float64
	Data        strategy.StrategyParams
	FromDt      int64
	ToDt        int64
}

// strategyRepository implements StrategyRepository interface
type strategyRepository struct {
	db  *sqlx.DB
	log *zap.Logger
}

// NewStrategyRepository creates a new instance of StrategyRepository
func NewStrategyRepository(db *sqlx.DB, log *zap.Logger) StrategyRepository {
	return &strategyRepository{
		db:  db,
		log: log.Named("strategy_repository"),
	}
}

func (r *strategyRepository) Fetch(
	ctx context.Context,
	id *int64,
	symbol *string,
	timeframe *string,
	dt_left *time.Time,
	dt_right *time.Time,
) ([]Strategy, error) {
	var strategies []Strategy
	query := `
		SELECT
			id,
			symbol,
			timeframe,
			trials,
			workers,
			val_set_days,
			train_set_day,
			win_rate,
			data,
			from_dt,
			to_dt,
			dt
		FROM strategy
		WHERE TRUE
			and ($1::bigint is null or id = $1)
			and ($2::text is null or symbol = $2)
			and ($3::text is null or timeframe = $3)
			and (
				($4::timestamp without time zone is null and $5::timestamp without time zone is null)
				or dt between $4 and $5
			)
		ORDER BY id
	`

	err := r.db.SelectContext(ctx, &strategies, query, id, symbol, timeframe, dt_left, dt_right)
	if err != nil {
		r.log.Error("failed to fetch strategies",
			zap.Error(err),
			zap.Any("id", id),
			zap.Any("symbol", symbol),
			zap.Any("timeframe", timeframe))
		return nil, err
	}

	return strategies, nil
}

// Create adds a new strategy to the database
func (r *strategyRepository) Create(ctx context.Context, strategy StrategyCreate) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO strategy (symbol, timeframe, trials, workers, val_set_days, train_set_day, 
							  win_rate, data, from_dt, to_dt, dt)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`,
		strategy.Symbol,
		strategy.Timeframe,
		strategy.Trials,
		strategy.Workers,
		strategy.ValSetDays,
		strategy.TrainSetDay,
		strategy.WinRate,
		StrategyData(strategy.Data),
		strategy.FromDt,
		strategy.ToDt,
		time.Now()).Scan(&id)

	if err != nil {
		r.log.Error("failed to create strategy",
			zap.String("symbol", strategy.Symbol),
			zap.String("timeframe", strategy.Timeframe),
			zap.Error(err))
		return 0, err
	}
	return id, nil
}

// Update modifies an existing strategy
func (r *strategyRepository) Update(ctx context.Context, id int64, data strategy.StrategyParams) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE strategy
		SET data = $1, dt = $2
		WHERE id = $3
	`, StrategyData(data), time.Now(), id)

	if err != nil {
		r.log.Error("failed to update strategy", zap.Int64("id", id), zap.Error(err))
		return err
	}
	return nil
}

// Delete removes a strategy from the database
func (r *strategyRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM strategy WHERE id = $1`, id)
	if err != nil {
		r.log.Error("failed to delete strategy", zap.Int64("id", id), zap.Error(err))
		return err
	}
	return nil
}

// FetchStrategyAcc retrieves strategy_acc records with flexible filtering
func (r *strategyRepository) FetchStrategyAcc(
	ctx context.Context,
	id *int64,
	strategyID *int64,
	symbol *string,
	dt_upd_left *time.Time,
	dt_upd_right *time.Time,
	dt_create_left *time.Time,
	dt_create_right *time.Time,
) ([]StrategyAcc, error) {
	var strategies []StrategyAcc

	query := `
		SELECT
			sa.id,
			sa.strategy_id,
			sa.symbol,
			sa.timeframe,
			sa.dt_upd,
			sa.dt_create
		FROM strategy_acc sa
		WHERE TRUE
			and ($1::bigint is null or sa.id = $1)
			and ($2::bigint is null or sa.strategy_id = $2)
			and ($3::text is null or sa.symbol = $3)
			and (
				($4::timestamp without time zone is null and $5::timestamp without time zone is null)
				or sa.dt_upd between $4 and $5
			)
			and (
				($6::timestamp without time zone is null and $7::timestamp without time zone is null)
				or sa.dt_create between $6 and $7
			)
		ORDER BY id
	`

	err := r.db.SelectContext(
		ctx,
		&strategies,
		query,
		id,
		strategyID,
		symbol,
		dt_upd_left,
		dt_upd_right,
		dt_create_left,
		dt_create_right,
	)

	if err != nil {
		r.log.Error("failed to fetch strategy_acc records",
			zap.Error(err),
			zap.Any("id", id),
			zap.Any("strategy_id", strategyID),
			zap.Any("symbol", symbol))
		return nil, err
	}

	return strategies, nil
}

// CreateStrategyAcc adds a new record to the strategy_acc table
func (r *strategyRepository) CreateStrategyAcc(ctx context.Context, strategyID int64) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, `
		with cte as (
			select
				s.symbol,
				s.timeframe
			from strategy s
			where true
				and s.id = $1
			limit 1
		)
		INSERT INTO strategy_acc (strategy_id, symbol, timeframe, dt_upd, dt_create)
		VALUES ($1, (select symbol from cte), (select timeframe from cte), now(), now())
		RETURNING id
	`, strategyID).Scan(&id)

	if err != nil {
		r.log.Error("failed to create strategy_acc",
			zap.Int64("strategyID", strategyID),
			zap.Error(err))
		return 0, err
	}
	return id, nil
}

// UpdateStrategyAcc updates a strategy_acc record
func (r *strategyRepository) UpdateStrategyAcc(ctx context.Context, id int64, strategyID int64) error {
	_, err := r.db.ExecContext(ctx, `
		with cte as (
			select
				s.symbol,
				s.timeframe
			from strategy s
			where true
				and s.id = $1
			limit 1
		)
		UPDATE strategy_acc
		SET
			strategy_id = $1,
			symbol = (select symbol from cte),
			timeframe = (select timeframe from cte),
			dt_upd = now()
		WHERE id = $2
	`, strategyID, id)

	if err != nil {
		r.log.Error("failed to update strategy_acc",
			zap.Int64("id", id),
			zap.Int64("strategyID", strategyID),
			zap.Error(err))
		return err
	}
	return nil
}

// DeleteStrategyAcc removes a record from the strategy_acc table
func (r *strategyRepository) DeleteStrategyAcc(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM strategy_acc WHERE id = $1`, id)
	if err != nil {
		r.log.Error("failed to delete strategy_acc", zap.Int64("id", id), zap.Error(err))
		return err
	}
	return nil
}
