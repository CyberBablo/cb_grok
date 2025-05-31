package main

import (
	"cb_grok/config"
	"cb_grok/internal/strategy/repository"
	"cb_grok/internal/utils/logger"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// CLI parameters
var (
	symbol     string
	dateStr    string
	timeFormat = "2006-01-02"
)

func init() {
	flag.StringVar(&symbol, "symbol", "", "Symbol to compare strategies for (e.g. BTCUSDT)")
	flag.StringVar(&dateStr, "dt", "", "Date to compare strategies (format: YYYY-MM-DD)")
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

	// Use current date if not provided
	if dateStr == "" {
		dateStr = time.Now().Format(timeFormat)
	}

	// Validate date format
	_, err := time.Parse(timeFormat, dateStr)
	if err != nil {
		fmt.Printf("Error: invalid date format. Please use %s format\n", timeFormat)
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
			runCompare(log, db, repo, symbol, dateStr)
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

func runCompare(
	log *zap.Logger,
	db *sqlx.DB,
	strategyRepo repository.StrategyRepository,
	symbol string,
	dateStr string,
) {
	log.Info("Starting strategy comparison",
		zap.String("symbol", symbol),
		zap.String("date", dateStr))

	ctx := context.Background()

	// Convert the provided date string to a time.Time
	targetDate, _ := time.Parse(timeFormat, dateStr)

	// Create time range for the entire day
	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour).Add(-time.Nanosecond)

	log.Info("Date range for search",
		zap.Time("start", startOfDay),
		zap.Time("end", endOfDay))

	// Step 1: Find all strategies for the given symbol
	symbolStrategies, err := strategyRepo.Fetch(ctx, nil, &symbol, nil, nil, nil)
	if err != nil {
		log.Error("Failed to fetch strategies for symbol",
			zap.String("symbol", symbol),
			zap.Error(err))
		return
	}

	if len(symbolStrategies) == 0 {
		log.Info("No strategies found for the symbol", zap.String("symbol", symbol))
		return
	}

	log.Info("Found strategies for symbol",
		zap.String("symbol", symbol),
		zap.Int("count", len(symbolStrategies)))

	// Step 2: Get current active strategies for this symbol
	symbolAccs, err := strategyRepo.FetchStrategyAcc(
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

	// Step 3: Find the best strategy (highest win_rate) for the given symbol
	var bestStrategy *repository.Strategy
	for _, s := range symbolStrategies {
		// Skip strategies created after the specified date
		if s.DT.After(endOfDay) {
			continue
		}

		// Initialize bestStrategy if it's the first valid strategy
		if bestStrategy == nil {
			bestStrategy = &s
			continue
		}

		// Update bestStrategy if current strategy has higher win_rate
		if s.WinRate > bestStrategy.WinRate {
			bestStrategy = &s
		}
	}

	if bestStrategy == nil {
		log.Info("No valid strategies found for the date",
			zap.String("symbol", symbol),
			zap.String("date", dateStr))
		return
	}

	log.Info("Found best strategy",
		zap.Int64("id", bestStrategy.ID),
		zap.String("symbol", bestStrategy.Symbol),
		zap.String("timeframe", bestStrategy.Timeframe),
		zap.Float64("win_rate", bestStrategy.WinRate))

	// Step 4: Update or create strategy_acc record
	if len(symbolAccs) > 0 {
		// There's already a strategy for this symbol
		acc := symbolAccs[0]

		// Check if we already have the best strategy
		if acc.StrategyID == bestStrategy.ID {
			log.Info("Best strategy is already set",
				zap.Int64("strategy_id", bestStrategy.ID),
				zap.Float64("win_rate", bestStrategy.WinRate))
			return
		}

		// Get details of the current strategy
		currentStrategy, err := strategyRepo.Fetch(ctx, &acc.StrategyID, nil, nil, nil, nil)
		if err != nil || len(currentStrategy) == 0 {
			log.Error("Failed to fetch current strategy",
				zap.Int64("id", acc.StrategyID),
				zap.Error(err))
			return
		}

		currentWinRate := currentStrategy[0].WinRate

		// Compare win rates
		if bestStrategy.WinRate > currentWinRate {
			log.Info("Found better strategy, updating record",
				zap.Float64("current_win_rate", currentWinRate),
				zap.Float64("new_win_rate", bestStrategy.WinRate),
				zap.Int64("old_strategy_id", acc.StrategyID),
				zap.Int64("new_strategy_id", bestStrategy.ID))

			// Update the strategy_acc record with the new best strategy
			err := strategyRepo.UpdateStrategyAcc(
				ctx,
				acc.ID,
				bestStrategy.ID,
				bestStrategy.Symbol,
				bestStrategy.Timeframe,
			)

			if err != nil {
				log.Error("Failed to update strategy", zap.Error(err))
				return
			}

			log.Info("Successfully updated strategy",
				zap.Int64("acc_id", acc.ID),
				zap.Int64("new_strategy_id", bestStrategy.ID))
		} else {
			log.Info("Current strategy has better or equal win_rate, no update needed",
				zap.Float64("current_win_rate", currentWinRate),
				zap.Float64("best_found_win_rate", bestStrategy.WinRate))
		}
	} else {
		// No strategy exists for this symbol, create a new one
		log.Info("No strategy found for symbol, creating new record",
			zap.String("symbol", symbol),
			zap.Int64("strategy_id", bestStrategy.ID))

		// Create a new strategy_acc record
		accID, err := strategyRepo.CreateStrategyAcc(
			ctx,
			bestStrategy.ID,
			bestStrategy.Symbol,
			bestStrategy.Timeframe,
		)
		if err != nil {
			log.Error("Failed to create strategy_acc record", zap.Error(err))
			return
		}

		log.Info("Successfully created new strategy record",
			zap.Int64("acc_id", accID),
			zap.Int64("strategy_id", bestStrategy.ID),
			zap.Float64("win_rate", bestStrategy.WinRate))
	}

	// Print summary information
	log.Info("Strategy comparison completed",
		zap.String("symbol", symbol),
		zap.String("date", dateStr))

	// Get all strategies after possible changes
	updatedAccs, _ := strategyRepo.FetchStrategyAcc(ctx, nil, nil, &symbol, nil, nil, nil, nil)

	fmt.Println("\nStrategies Summary:")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("%-10s %-10s %-12s %-10s %-12s\n",
		"Acc ID", "Strategy ID", "Symbol", "Timeframe", "Win Rate")
	fmt.Println(strings.Repeat("-", 80))

	for _, acc := range updatedAccs {
		// Get strategy details
		strategies, _ := strategyRepo.Fetch(ctx, &acc.StrategyID, nil, nil, nil, nil)
		if len(strategies) > 0 {
			s := strategies[0]
			fmt.Printf("%-10d %-10d %-12s %-10s %-12.2f%%\n",
				acc.ID, s.ID, acc.Symbol, acc.Timeframe, s.WinRate*100)
		}
	}
	fmt.Println(strings.Repeat("-", 80))

	return
}
