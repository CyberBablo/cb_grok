package main

import (
	"cb_grok/config"
	"cb_grok/internal/strategy"
	"cb_grok/internal/strategy/repository"
	"cb_grok/internal/utils/logger"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// CLI parameters
var (
	symbolsCount int
	count        int
	clearData    bool
)

// List of test symbols
var symbols = []string{
	"BTCUSDT", "ETHUSDT", "BNBUSDT", "XRPUSDT", "ADAUSDT",
	"SOLUSDT", "DOTUSDT", "DOGEUSDT", "AVAXUSDT", "SHIBUSDT",
}

// List of test timeframes
var timeframes = []string{
	"1m", "5m", "15m", "30m", "1h", "4h", "1d", "1w",
}

func init() {
	flag.IntVar(&symbolsCount, "symbols", 5, "Number of different symbols to generate (max 10)")
	flag.IntVar(&count, "count", 50, "Number of strategy records to generate")
	flag.BoolVar(&clearData, "clear", false, "Clear existing data before generating new data")
}

func main() {
	// Parse command line arguments
	flag.Parse()

	// Validate parameters
	if symbolsCount <= 0 || symbolsCount > len(symbols) {
		fmt.Printf("Invalid symbol count. Must be between 1 and %d\n", len(symbols))
		os.Exit(1)
	}

	if count <= 0 {
		fmt.Println("Count must be greater than 0")
		os.Exit(1)
	}

	// Create and run fx application
	fx.New(
		logger.Module,
		config.Module,
		providePgxConnection(),
		repository.Module,
		fx.Invoke(func(log *zap.Logger, db *sqlx.DB, repo repository.StrategyRepository) {
			generateTestData(log, db, repo)
		}),
	).Run()
}

func providePgxConnection() fx.Option {
	return fx.Provide(func(log *zap.Logger) *sqlx.DB {
		// DSN: формат подключения
		dsn := "postgres://postgres:password@0.0.0.0:5433/cb_grok?sslmode=disable"

		// Используем pgx через stdlib
		db := sqlx.MustConnect("pgx", dsn)
		log.Info("Connected to database")

		return db
	})
}

func generateTestData(
	log *zap.Logger,
	db *sqlx.DB,
	strategyRepo repository.StrategyRepository,
) {
	log.Info("Starting test data generation",
		zap.Int("symbols", symbolsCount),
		zap.Int("count", count),
		zap.Bool("clear_data", clearData))

	ctx := context.Background()

	// In newer Go versions, rand is auto-seeded, but we'll keep this for compatibility
	// with older Go versions
	rand.Seed(time.Now().UnixNano())

	// Clear existing data if requested
	if clearData {
		log.Info("Clearing existing data")

		// First delete strategy_acc records
		_, err := db.ExecContext(ctx, "DELETE FROM strategy_acc")
		if err != nil {
			log.Error("Failed to clear strategy_acc data", zap.Error(err))
			return
		}

		// Then delete strategy records
		_, err = db.ExecContext(ctx, "DELETE FROM strategy")
		if err != nil {
			log.Error("Failed to clear strategy data", zap.Error(err))
			return
		}

		// Clear strategy_timeframe records
		_, err = db.ExecContext(ctx, "DELETE FROM strategy_timeframe")
		if err != nil {
			log.Error("Failed to clear strategy_timeframe data", zap.Error(err))
			return
		}

		log.Info("Existing data cleared")
	}

	// Insert or update timeframe configurations
	log.Info("Setting up timeframe configurations")
	for _, tf := range timeframes {
		// Generate random dt_diff values (in seconds) between 1 day and 7 days
		dtDiff := int64(86400 + rand.Intn(6*86400)) // 1-7 days in seconds

		// Check if entry exists
		var count int
		err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM strategy_timeframe WHERE name = $1", tf)
		if err != nil {
			log.Error("Failed to check strategy_timeframe", zap.Error(err))
			continue
		}

		if count == 0 {
			// Insert new record
			_, err := db.ExecContext(ctx, `
				INSERT INTO strategy_timeframe (name, value)
				VALUES ($1, $2)
			`, tf, fmt.Sprintf(`{"dt_diff":%d}`, dtDiff))

			if err != nil {
				log.Error("Failed to insert strategy_timeframe",
					zap.String("timeframe", tf),
					zap.Error(err))
			} else {
				log.Info("Created timeframe config",
					zap.String("timeframe", tf),
					zap.Int64("dt_diff", dtDiff))
			}
		} else {
			// Update existing record
			_, err := db.ExecContext(ctx, `
				UPDATE strategy_timeframe
				SET value = $2
				WHERE name = $1
			`, tf, fmt.Sprintf(`{"dt_diff":%d}`, dtDiff))

			if err != nil {
				log.Error("Failed to update strategy_timeframe",
					zap.String("timeframe", tf),
					zap.Error(err))
			} else {
				log.Info("Updated timeframe config",
					zap.String("timeframe", tf),
					zap.Int64("dt_diff", dtDiff))
			}
		}
	}

	// Select subset of symbols to use
	selectedSymbols := symbols[:symbolsCount]

	// Generate strategy records
	log.Info("Generating strategy records")

	strategyIDs := make(map[string][]int64)

	for i := 0; i < count; i++ {
		// Select random symbol and timeframe
		symbol := selectedSymbols[rand.Intn(len(selectedSymbols))]
		timeframe := timeframes[rand.Intn(len(timeframes))]

		// Generate random win rate between 0.3 and 0.9
		winRate := 0.3 + rand.Float64()*0.6

		// Generate random strategy parameters
		strategyParams := generateRandomStrategyParams()

		// Generate random date within the last 30 days
		now := time.Now()
		daysAgo := rand.Intn(30)
		creationDate := now.AddDate(0, 0, -daysAgo)

		// Convert to Unix timestamp
		fromDt := creationDate.AddDate(0, -3, 0).Unix() // 3 months before creation date
		toDt := creationDate.Unix()

		// Create random trials and workers count
		trials := 50 + rand.Intn(951) // 50 to 1000
		workers := 1 + rand.Intn(8)   // 1 to 8

		// Random validation and training days
		valDays := 3 + rand.Intn(8)     // 3 to 10
		trainDays := 10 + rand.Intn(21) // 10 to 30

		// Create the strategy
		createParams := repository.StrategyCreate{
			Symbol:      symbol,
			Timeframe:   timeframe,
			Trials:      trials,
			Workers:     workers,
			ValSetDays:  valDays,
			TrainSetDay: trainDays,
			WinRate:     winRate,
			Data:        strategyParams,
			FromDt:      fromDt,
			ToDt:        toDt,
		}

		id, err := strategyRepo.Create(ctx, createParams)
		if err != nil {
			log.Error("Failed to create strategy", zap.Error(err))
			continue
		}

		// Store ID for later use
		key := fmt.Sprintf("%s_%s", symbol, timeframe)
		strategyIDs[key] = append(strategyIDs[key], id)

		// Log progress every 10 records
		if (i+1)%10 == 0 || i+1 == count {
			log.Info("Progress",
				zap.Int("created", i+1),
				zap.Int("total", count))
		}
	}

	// Create some strategy_acc records (one for each symbol)
	log.Info("Creating strategy_acc records")

	for _, symbol := range selectedSymbols {
		// Find the best strategy for this symbol
		var bestID int64
		var bestWinRate float64
		var bestTimeframe string

		for key, ids := range strategyIDs {
			if len(ids) == 0 {
				continue
			}

			// Check if this key has the current symbol
			if key[:len(symbol)] == symbol {
				// Fetch each strategy and find the one with highest win rate
				for _, id := range ids {
					strategies, err := strategyRepo.Fetch(ctx, &id, nil, nil, nil, nil, nil)
					if err != nil || len(strategies) == 0 {
						continue
					}

					s := strategies[0]
					if s.WinRate > bestWinRate {
						bestWinRate = s.WinRate
						bestID = s.ID
						bestTimeframe = s.Timeframe
					}
				}
			}
		}

		if bestID != 0 {
			// Create a strategy_acc record for the best strategy
			accID, err := strategyRepo.CreateStrategyAcc(ctx, bestID)
			if err != nil {
				log.Error("Failed to create strategy_acc",
					zap.String("symbol", symbol),
					zap.Error(err))
				continue
			}

			log.Info("Created strategy_acc record",
				zap.String("symbol", symbol),
				zap.String("timeframe", bestTimeframe),
				zap.Int64("strategy_id", bestID),
				zap.Float64("win_rate", bestWinRate),
				zap.Int64("acc_id", accID))
		}
	}

	// Print summary
	log.Info("Data generation completed",
		zap.Int("strategies_created", count))

	// Count strategy_acc records
	var accCount int
	err := db.GetContext(ctx, &accCount, "SELECT COUNT(*) FROM strategy_acc")
	if err == nil {
		log.Info("Strategy_acc records created", zap.Int("count", accCount))
	}

	// Count by symbol
	fmt.Println("\nStrategies by Symbol:")
	for _, symbol := range selectedSymbols {
		var symbolCount int
		symParam := symbol
		err := db.GetContext(ctx, &symbolCount, "SELECT COUNT(*) FROM strategy WHERE symbol = $1", symParam)
		if err == nil {
			fmt.Printf("%-10s: %d strategies\n", symbol, symbolCount)
		}
	}

	return
}

func generateRandomStrategyParams() strategy.StrategyParams {
	return strategy.StrategyParams{
		MAShortPeriod:       5 + rand.Intn(15),         // 5-20
		MALongPeriod:        20 + rand.Intn(80),        // 20-100
		RSIPeriod:           5 + rand.Intn(20),         // 5-25
		BuyRSIThreshold:     20 + rand.Float64()*20,    // 20-40
		SellRSIThreshold:    60 + rand.Float64()*20,    // 60-80
		EMAShortPeriod:      3 + rand.Intn(17),         // 3-20
		EMALongPeriod:       20 + rand.Intn(80),        // 20-100
		ATRPeriod:           5 + rand.Intn(20),         // 5-25
		ATRThreshold:        0.5 + rand.Float64()*2.5,  // 0.5-3.0
		MACDShortPeriod:     8 + rand.Intn(8),          // 8-16
		MACDLongPeriod:      20 + rand.Intn(10),        // 20-30
		MACDSignalPeriod:    5 + rand.Intn(10),         // 5-15
		EMAWeight:           0.5 + rand.Float64()*1.5,  // 0.5-2.0
		TrendWeight:         0.5 + rand.Float64()*1.5,  // 0.5-2.0
		RSIWeight:           0.5 + rand.Float64()*1.5,  // 0.5-2.0
		MACDWeight:          0.5 + rand.Float64()*1.5,  // 0.5-2.0
		BuySignalThreshold:  0.3 + rand.Float64()*0.4,  // 0.3-0.7
		SellSignalThreshold: -0.7 + rand.Float64()*0.4, // -0.7 to -0.3
		BollingerPeriod:     10 + rand.Intn(30),        // 10-40
		BollingerStdDev:     1.5 + rand.Float64()*1.5,  // 1.5-3.0
		BBWeight:            0.5 + rand.Float64()*1.5,  // 0.5-2.0
		StochasticKPeriod:   5 + rand.Intn(15),         // 5-20
		StochasticDPeriod:   2 + rand.Intn(4),          // 2-6
		StochasticWeight:    0.5 + rand.Float64()*1.5,  // 0.5-2.0
	}
}
