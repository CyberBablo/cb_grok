// Package metrics provides functionality to generate realistic trading metrics data
// for testing and development purposes.
package metrics

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"cb_grok/config"
	"cb_grok/internal/metrics"
	"cb_grok/internal/metrics/repository"
	metrics_model "cb_grok/internal/models/metrics"
	"cb_grok/pkg/logger"
	"cb_grok/pkg/postgres"
)

// MetricsParams holds the command line parameters for metrics operations.
type MetricsParams struct {
	Symbol         string
	Days           int
	Trades         int
	Environment    string
	StrategyType   string
	InitialCapital float64
}

// CMD runs the metrics generation command with FX dependency injection.
func CMD() {
	// Parse command line flags
	params := parseMetricsFlags()

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create FX application
	app := fx.New(
		// Provide configuration
		fx.Provide(func() *config.Config { return cfg }),
		fx.Provide(func() MetricsParams { return params }),

		// Provide logger
		fx.Provide(func(cfg *config.Config) (*zap.Logger, error) {
			return logger.NewZapLogger(logger.ZapConfig{
				Level:        cfg.Logger.Level,
				Development:  cfg.Logger.Development,
				Encoding:     cfg.Logger.Encoding,
				OutputPaths:  cfg.Logger.OutputPaths,
				FileLog:      cfg.Logger.FileLog,
				FilePath:     cfg.Logger.FilePath,
				FileMaxSize:  cfg.Logger.FileMaxSize,
				FileCompress: cfg.Logger.FileCompress,
			})
		}),

		// Provide database connection
		fx.Provide(func(cfg *config.Config) (postgres.Postgres, error) {
			return postgres.InitPsqlDB(&postgres.Conn{
				Host:     cfg.Postgres.Host,
				Port:     cfg.Postgres.Port,
				User:     cfg.Postgres.User,
				Password: cfg.Postgres.Password,
				DBName:   cfg.Postgres.DBName,
				SSLMode:  cfg.Postgres.SSLMode,
				PgDriver: cfg.Postgres.PgDriver,
			})
		}),

		// Provide metrics repository
		fx.Provide(func(db postgres.Postgres) metrics.Repository {
			return repository.New(db)
		}),

		// Lifecycle management
		fx.Invoke(runMetricsOperation),

		// FX logger
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

func populateMetrics(repo metrics.Repository, symbol string, days, trades int, logger *zap.Logger) error {
	rand.Seed(time.Now().UnixNano())

	// Create strategy run
	strategyParams, _ := json.Marshal(map[string]interface{}{
		"type":           "demo_strategy",
		"short_ma":       20,
		"long_ma":        50,
		"rsi_oversold":   30,
		"rsi_overbought": 70,
	})

	initialCapital := 10000.0
	runID := uuid.New().String()
	run := &metrics_model.StrategyRun{
		RunID:          runID,
		Symbol:         symbol,
		StartTime:      time.Now().AddDate(0, 0, -days),
		InitialCapital: initialCapital,
		StrategyType:   "demo",
		StrategyParams: strategyParams,
		Environment:    "demo",
	}

	err := repo.CreateStrategyRun(context.Background(), run)
	if err != nil {
		return fmt.Errorf("failed to create strategy run: %w", err)
	}

	logger.Info("Created strategy run", zap.String("run_id", runID))

	// Generate time series data (every hour for the specified days)
	startTime := time.Now().AddDate(0, 0, -days)
	endTime := time.Now()

	portfolioValue := initialCapital
	btcPrice := 45000.0 // Starting BTC price

	logger.Info("Generating time series data...")

	for current := startTime; current.Before(endTime); current = current.Add(time.Hour) {
		// Simulate price movement
		btcPrice += (rand.Float64() - 0.5) * btcPrice * 0.02 // ±2% hourly movement
		if btcPrice < 20000 {
			btcPrice = 20000
		}
		if btcPrice > 80000 {
			btcPrice = 80000
		}

		// Generate realistic indicators
		indicators := generateIndicators(btcPrice, current)

		// Update portfolio value with some volatility
		portfolioValue += (rand.Float64() - 0.5) * portfolioValue * 0.01 // ±1% hourly movement

		// Save portfolio value
		err := repo.SaveTimeSeriesMetric(current, symbol, "portfolio_value", portfolioValue, nil)
		if err != nil {
			return fmt.Errorf("failed to save portfolio value: %w", err)
		}

		// Save indicators
		for name, value := range indicators {
			err := repo.SaveTimeSeriesMetric(
				current,
				symbol,
				"indicator_"+name,
				value,
				map[string]interface{}{"indicator": name},
			)
			if err != nil {
				return fmt.Errorf("failed to save indicator %s: %w", name, err)
			}
		}

		// Save performance metrics every 24 hours
		if current.Hour() == 0 {
			winRate := 45 + rand.Float64()*20   // 45-65%
			maxDrawdown := rand.Float64() * 15  // 0-15%
			sharpeRatio := 0.5 + rand.Float64() // 0.5-1.5

			metrics := map[string]float64{
				"win_rate":     winRate,
				"max_drawdown": maxDrawdown,
				"sharpe_ratio": sharpeRatio,
			}

			for name, value := range metrics {
				err := repo.SaveTimeSeriesMetric(current, symbol, name, value, nil)
				if err != nil {
					return fmt.Errorf("failed to save metric %s: %w", name, err)
				}
			}
		}
	}

	// Generate trade data
	logger.Info("Generating trade data...")

	tradeStartTime := startTime
	position := 0.0 // 0 = no position, positive = long position
	wins := 0
	losses := 0
	totalProfit := 0.0

	for i := 0; i < trades; i++ {
		// Generate trade timestamp
		tradeTime := tradeStartTime.Add(time.Duration(rand.Intn(int(endTime.Sub(tradeStartTime).Hours()))) * time.Hour)

		// Generate trade data
		price := 40000 + rand.Float64()*20000 // $40k-$60k range
		quantity := 0.1 + rand.Float64()*0.4  // 0.1-0.5 BTC

		var side string
		var profit float64

		if position == 0 {
			// Enter position (buy)
			side = "buy"
			position = quantity
			profit = 0
		} else {
			// Exit position (sell)
			side = "sell"
			// Calculate profit based on price movement
			entryPrice := price * (0.95 + rand.Float64()*0.1) // Entry was within ±5% of current
			profit = (price - entryPrice) * position
			totalProfit += profit

			if profit > 0 {
				wins++
			} else {
				losses++
			}
			position = 0
		}

		// Generate indicators for this trade
		indicators := generateIndicators(price, tradeTime)
		indicatorsJSON, _ := json.Marshal(indicators)

		// Calculate current performance metrics
		currentWinRate := 0.0
		if wins+losses > 0 {
			currentWinRate = float64(wins) / float64(wins+losses) * 100
		}

		currentDrawdown := rand.Float64() * 10 // 0-10%
		currentSharpe := 0.5 + rand.Float64()  // 0.5-1.5

		// Determine decision trigger
		triggers := []string{"signal", "stop_loss", "take_profit"}
		trigger := triggers[rand.Intn(len(triggers))]

		trade := &metrics_model.TradeMetric{
			Timestamp:       tradeTime,
			Symbol:          symbol,
			Side:            side,
			Price:           price,
			Quantity:        quantity,
			Profit:          &profit,
			PortfolioValue:  &portfolioValue,
			Indicators:      indicatorsJSON,
			DecisionTrigger: &trigger,
			WinRate:         &currentWinRate,
			MaxDrawdown:     &currentDrawdown,
			SharpeRatio:     &currentSharpe,
		}

		err := repo.SaveTradeMetric(context.Background(), trade)
		if err != nil {
			return fmt.Errorf("failed to save trade %d: %w", i+1, err)
		}

		if i%20 == 0 {
			logger.Info("Generated trades", zap.Int("count", i+1), zap.Int("total", trades))
		}
	}

	// Update strategy run with final results
	finalCapital := initialCapital + totalProfit
	finalWinRate := 0.0
	if wins+losses > 0 {
		finalWinRate = float64(wins) / float64(wins+losses) * 100
	}

	maxDrawdown := 5 + rand.Float64()*10    // 5-15%
	sharpeRatio := 0.8 + rand.Float64()*0.7 // 0.8-1.5
	endTimePtr := time.Now()

	finalRun := &metrics_model.StrategyRun{
		RunID:         runID,
		EndTime:       &endTimePtr,
		FinalCapital:  &finalCapital,
		TotalTrades:   trades,
		WinningTrades: wins,
		LosingTrades:  losses,
		TotalProfit:   totalProfit,
		MaxDrawdown:   &maxDrawdown,
		SharpeRatio:   &sharpeRatio,
		WinRate:       &finalWinRate,
	}

	err = repo.UpdateStrategyRun(context.Background(), finalRun)
	if err != nil {
		return fmt.Errorf("failed to update strategy run: %w", err)
	}

	logger.Info("Population summary",
		zap.String("symbol", symbol),
		zap.Int("total_trades", trades),
		zap.Int("winning_trades", wins),
		zap.Int("losing_trades", losses),
		zap.Float64("total_profit", totalProfit),
		zap.Float64("win_rate", finalWinRate),
		zap.Float64("final_capital", finalCapital))

	return nil
}

func generateIndicators(price float64, timestamp time.Time) map[string]float64 {
	// Generate realistic technical indicators
	hour := float64(timestamp.Hour())
	day := float64(timestamp.YearDay())

	// Use sine waves and price for more realistic indicators
	priceNorm := (price - 40000) / 20000 // Normalize price to -1 to 1 range

	return map[string]float64{
		"RSI":         30 + 40*(0.5+0.5*math.Sin(day/10+priceNorm)),    // 30-70 range
		"ATR":         price * (0.02 + 0.01*math.Abs(math.Sin(day/5))), // 2-3% of price
		"MACD":        price * 0.001 * math.Sin(day/7+priceNorm),       // Oscillating around 0
		"ADX":         20 + 30*(0.5+0.5*math.Sin(day/12)),              // 20-50 range
		"StochasticK": 20 + 60*(0.5+0.5*math.Sin(hour/6+priceNorm)),    // 20-80 range
		"StochasticD": 25 + 50*(0.5+0.5*math.Sin(hour/8+priceNorm)),    // 25-75 range
		"ShortMA":     price * (0.99 + 0.02*math.Sin(day/3)),           // Close to current price
		"LongMA":      price * (0.97 + 0.06*math.Sin(day/15)),          // More stable
		"ShortEMA":    price * (0.995 + 0.01*math.Sin(day/2)),          // Very close to price
		"LongEMA":     price * (0.96 + 0.08*math.Sin(day/20)),          // More stable
		"UpperBB":     price * (1.02 + 0.01*math.Abs(math.Sin(day/8))), // Above price
		"LowerBB":     price * (0.98 - 0.01*math.Abs(math.Sin(day/8))), // Below price
	}
}

// parseMetricsFlags parses command line flags and returns MetricsParams.
func parseMetricsFlags() MetricsParams {
	var (
		symbol   = flag.String("symbol", "BTCUSDT", "Trading symbol (e.g., BTCUSDT, ETHUSDT)")
		days     = flag.Int("days", 30, "Number of days to generate historical data for")
		trades   = flag.Int("trades", 100, "Number of trades to generate")
		env      = flag.String("env", "demo", "Environment for the strategy run (demo, backtest, live)")
		strategy = flag.String("strategy", "demo_strategy", "Strategy type name")
		capital  = flag.Float64("capital", 10000.0, "Initial capital amount")
	)
	flag.Parse()

	return MetricsParams{
		Symbol:         *symbol,
		Days:           *days,
		Trades:         *trades,
		Environment:    *env,
		StrategyType:   *strategy,
		InitialCapital: *capital,
	}
}

// runMetricsOperation is the main FX lifecycle function that executes metrics operations.
func runMetricsOperation(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	metricsRepo metrics.Repository,
	params MetricsParams,
	shutdowner fx.Shutdowner,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting metrics generation with FX",
				zap.String("symbol", params.Symbol),
				zap.Int("days", params.Days),
				zap.Int("trades", params.Trades),
				zap.String("environment", params.Environment),
				zap.String("strategy", params.StrategyType),
				zap.Float64("initial_capital", params.InitialCapital))

			go func() {
				defer shutdowner.Shutdown()

				// Execute metrics generation
				err := populateMetrics(metricsRepo, params.Symbol, params.Days, params.Trades, log)
				if err != nil {
					log.Error("Metrics generation failed", zap.Error(err))
					return
				}

				log.Info("✅ Metrics generation completed successfully")
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Metrics generation completed")
			return nil
		},
	})
}

// MetricsGenerationParams holds parameters for metrics data generation.
type MetricsGenerationParams struct {
	Symbol         string
	Days           int
	Trades         int
	Environment    string
	StrategyType   string
	InitialCapital float64
}
