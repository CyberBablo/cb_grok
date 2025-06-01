package main

import (
	"cb_grok/config"
	"cb_grok/internal/strategy/repository"
	"cb_grok/internal/utils/logger"
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// CLI parameters
var (
	symbol     string
	timeFormat = "2006-01-02"
)

func init() {
	flag.StringVar(&symbol, "symbol", "", "Symbol to compare strategies for (e.g. BTCUSDT)")
}

func main() {
	// Parse CLI flags
	flag.Parse()

	// Validate parameters
	if symbol == "" {
		fmt.Println("Error: symbol parameter is required")
		flag.Usage()
		os.Exit(1)
	}

	// Create and run fx application
	fx.New(
		logger.Module,
		config.Module,
		providePgxConnection(),
		repository.Module,
		fx.Invoke(func(log *zap.Logger, db *sqlx.DB, repo repository.StrategyRepository) {
			runCompare(log, db, repo, symbol)
		}),
	).Run()
}

func providePgxConnection() fx.Option {
	return fx.Provide(func(log *zap.Logger, cfg config.Config) *sqlx.DB {
		dsn := cfg.PostgreSQL.DSN()
		log.Info("Connecting to database", zap.String("dsn", dsn))

		db := sqlx.MustConnect("pgx", dsn)
		log.Info("Connected to database")

		return db
	})
}

func logic(
	log *zap.Logger,
	db *sqlx.DB,
	strategyRepo repository.StrategyRepository,
	symbol string,
) {
	ctx := context.Background()

	log.Info("Starting strategy optimization process",
		zap.String("symbol", symbol),
		zap.Time("execution_time", time.Now()))

	dt_now := time.Now()
	log.Info("Time window for strategy search",
		zap.Time("current_time", dt_now))

	// Fetch active strategies with timeframe configuration
	log.Info("Fetching active strategies with timeframe configuration",
		zap.String("symbol", symbol))

	accStrategyArray, err := strategyRepo.FetchStrategyAccEx(
		ctx,
		nil,      // id
		nil,      // strategyID
		&symbol,  // symbol
		nil, nil, // dt_upd range
		nil, nil, // dt_create range
	)
	if err != nil {
		log.Error("Failed to fetch active strategies", zap.Error(err))
		return
	}

	log.Info("Found active strategies",
		zap.Int("count", len(accStrategyArray)),
		zap.String("symbol", symbol))
	for i, v := range accStrategyArray {
		log.Info("Processing strategy account",
			zap.Int("index", i),
			zap.Int64("acc_id", v.ID),
			zap.Int64("current_strategy_id", v.StrategyID),
			zap.String("symbol", v.Symbol),
			zap.String("timeframe", v.Timeframe),
			zap.Float64("current_win_rate", v.WinRate))

		// Convert DTDiff (unixtime) to a proper time.Time by calculating the date
		dt_from := dt_now.Add(time.Duration(-v.TimeframeConfig.DTDiff) * time.Second)

		log.Info("Searching for better strategies in time window",
			zap.Time("from", dt_from),
			zap.Time("to", dt_now),
			zap.Int64("time_window_seconds", v.TimeframeConfig.DTDiff))

		baseStrategyArray, err := strategyRepo.Fetch(
			ctx,
			nil,
			&symbol,
			&v.Timeframe,
			&dt_from,
			&dt_now,
			&v.WinRate,
		)
		if err != nil {
			log.Error("Failed to fetch strategies for symbol",
				zap.String("symbol", symbol),
				zap.String("timeframe", v.Timeframe),
				zap.Error(err))
			return
		}

		log.Info("Found candidate strategies",
			zap.Int("count", len(baseStrategyArray)),
			zap.String("timeframe", v.Timeframe))
		var bestStrategy *repository.Strategy

		// Find best strategy by win rate
		log.Info("Finding best strategy from candidates")
		for j, b := range baseStrategyArray {
			if bestStrategy == nil {
				log.Debug("Initial best strategy candidate",
					zap.Int("index", j),
					zap.Int64("id", b.ID),
					zap.Float64("win_rate", b.WinRate))
				bestStrategy = &b
				continue
			}
			if b.WinRate > bestStrategy.WinRate {
				log.Debug("Found better strategy",
					zap.Int("index", j),
					zap.Int64("id", b.ID),
					zap.Float64("win_rate", b.WinRate),
					zap.Float64("previous_best_win_rate", bestStrategy.WinRate))
				bestStrategy = &b
			}
		}

		// Check if we found any strategies
		if bestStrategy == nil {
			log.Info("No suitable strategies found for symbol and timeframe",
				zap.String("symbol", symbol),
				zap.String("timeframe", v.Timeframe))
			continue
		}

		log.Info("Found best strategy",
			zap.Int64("id", bestStrategy.ID),
			zap.String("timeframe", bestStrategy.Timeframe),
			zap.Float64("win_rate", bestStrategy.WinRate),
			zap.Time("strategy_date", bestStrategy.DT))

		// Compare with current strategy
		if bestStrategy.ID == v.StrategyID {
			log.Info("Best strategy is already assigned to this account - no update needed",
				zap.Int64("strategy_id", bestStrategy.ID),
				zap.Int64("account_id", v.ID))
			continue
		}

		// Only update if new strategy has better win rate
		if bestStrategy.WinRate <= v.WinRate {
			log.Info("Current strategy has better or equal win rate - no update needed",
				zap.Float64("current_win_rate", v.WinRate),
				zap.Float64("best_found_win_rate", bestStrategy.WinRate))
			continue
		}

		// Update the strategy account
		log.Info("Updating strategy account with better strategy",
			zap.Int64("account_id", v.ID),
			zap.Int64("old_strategy_id", v.StrategyID),
			zap.Int64("new_strategy_id", bestStrategy.ID),
			zap.Float64("old_win_rate", v.WinRate),
			zap.Float64("new_win_rate", bestStrategy.WinRate),
			zap.Float64("win_rate_improvement", bestStrategy.WinRate-v.WinRate))

		err = strategyRepo.UpdateStrategyAcc(
			ctx,
			v.ID,
			bestStrategy.ID,
		)
		if err != nil {
			log.Error("Failed to update strategy account",
				zap.Int64("account_id", v.ID),
				zap.Int64("strategy_id", bestStrategy.ID),
				zap.Error(err))
			return
		}
		log.Info("Successfully updated strategy account",
			zap.Int64("account_id", v.ID),
			zap.Int64("new_strategy_id", bestStrategy.ID))
	}
}

func runCompare(
	log *zap.Logger,
	db *sqlx.DB,
	strategyRepo repository.StrategyRepository,
	symbol string,
) {
	startTime := time.Now()
	log.Info("Starting strategy comparison process",
		zap.String("symbol", symbol),
		zap.Time("start_time", startTime))

	logic(log, db, strategyRepo, symbol)

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	log.Info("Strategy comparison process completed",
		zap.String("symbol", symbol),
		zap.Duration("duration", duration),
		zap.Time("end_time", endTime))
}
