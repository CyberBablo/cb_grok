// Package candle provides functionality to generate, query, and manage candle data
// for testing and development purposes.
package candle

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"cb_grok/config"
	"cb_grok/internal/candle"
	candleRepository "cb_grok/internal/candle/repository"
	candle_model "cb_grok/internal/models/candle"
	"cb_grok/pkg/logger"
	"cb_grok/pkg/postgres"
)

// CandleParams holds the command line parameters for candle operations.
type CandleParams struct {
	Operation string
	Symbol    string
	Exchange  string
	Timeframe string
	Count     int
}

// CMD runs the candle management command with FX dependency injection.
func CMD() {
	// Parse command line flags
	params := parseCandleFlags()

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
		fx.Provide(func() CandleParams { return params }),

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
				Host:     cfg.PostgresMetrics.Host,
				Port:     cfg.PostgresMetrics.Port,
				User:     cfg.PostgresMetrics.User,
				Password: cfg.PostgresMetrics.Password,
				DBName:   cfg.PostgresMetrics.DBName,
				SSLMode:  cfg.PostgresMetrics.SSLMode,
				PgDriver: cfg.PostgresMetrics.PgDriver,
			})
		}),

		// Provide candle repository
		fx.Provide(func(db postgres.Postgres) candle.Repository {
			return candleRepository.New(db)
		}),

		// Lifecycle management
		fx.Invoke(runCandleOperation),

		// FX logger
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

// parseCandleFlags parses command line flags and returns CandleParams.
func parseCandleFlags() CandleParams {
	var (
		operation = flag.String("op", "generate", "Operation: generate, query, delete")
		symbol    = flag.String("symbol", "BTCUSDT", "Trading symbol (e.g., BTCUSDT, ETHUSDT)")
		exchange  = flag.String("exchange", "bybit", "Exchange name")
		timeframe = flag.String("timeframe", "1m", "Timeframe (1m, 5m, 15m, 1h, etc.)")
		count     = flag.Int("count", 100, "Number of candles to generate/query")
	)
	flag.Parse()

	return CandleParams{
		Operation: *operation,
		Symbol:    *symbol,
		Exchange:  *exchange,
		Timeframe: *timeframe,
		Count:     *count,
	}
}

// runCandleOperation is the main FX lifecycle function that executes candle operations.
func runCandleOperation(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	candleRepo candle.Repository,
	params CandleParams,
	shutdowner fx.Shutdowner,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting candle operation with FX",
				zap.String("operation", params.Operation),
				zap.String("symbol", params.Symbol),
				zap.String("exchange", params.Exchange),
				zap.String("timeframe", params.Timeframe),
				zap.Int("count", params.Count))

			go func() {
				defer shutdowner.Shutdown()

				// Execute the requested operation
				var err error
				switch params.Operation {
				case "generate":
					err = generateCandles(ctx, candleRepo, params.Symbol, params.Exchange, params.Timeframe, params.Count, log)
				case "query":
					err = queryCandles(ctx, candleRepo, params.Symbol, params.Exchange, params.Timeframe, log)
				case "delete":
					err = deleteCandles(ctx, candleRepo, params.Symbol, params.Exchange, params.Timeframe, log)
				default:
					log.Error("Unknown operation", zap.String("operation", params.Operation))
					return
				}

				if err != nil {
					log.Error("Operation failed",
						zap.String("operation", params.Operation),
						zap.Error(err))
					return
				}

				log.Info("✅ Operation completed successfully",
					zap.String("operation", params.Operation))
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Candle operation completed")
			return nil
		},
	})
}

// generateCandles generates and saves realistic candle data.
func generateCandles(ctx context.Context, repo candle.Repository, symbol, exchange, timeframe string, count int, log *zap.Logger) error {
	log.Info("Generating candles",
		zap.String("symbol", symbol),
		zap.String("exchange", exchange),
		zap.String("timeframe", timeframe),
		zap.Int("count", count))

	// Start from current time and go backwards
	endTime := time.Now()
	interval := getTimeframeInterval(timeframe)

	// Generate realistic OHLCV data
	basePrice := getBasePriceForSymbol(symbol)
	rand.Seed(time.Now().UnixNano())

	successCount := 0
	for i := 0; i < count; i++ {
		// Calculate timestamp for this candle
		candleTime := endTime.Add(-time.Duration(count-i-1) * interval)
		timestamp := candleTime.Unix() * 1000 // Convert to milliseconds

		// Generate realistic OHLCV values
		candleData := generateRealisticOHLCV(basePrice, timestamp)

		// Save candle
		if err := repo.Create(ctx, symbol, exchange, timeframe, candleData); err != nil {
			log.Error("Failed to save candle",
				zap.Error(err),
				zap.Int("index", i),
				zap.Int64("timestamp", candleData.Timestamp))
			continue
		}

		successCount++
		basePrice = candleData.Close // Update base price for next candle

		if (i+1)%25 == 0 {
			log.Info("Generation progress", zap.Int("completed", i+1), zap.Int("total", count))
		}
	}

	log.Info("Candle generation completed",
		zap.Int("requested", count),
		zap.Int("saved", successCount))

	return nil
}

// queryCandles retrieves and displays candle data.
func queryCandles(ctx context.Context, repo candle.Repository, symbol, exchange, timeframe string, log *zap.Logger) error {
	log.Info("Querying candles",
		zap.String("symbol", symbol),
		zap.String("exchange", exchange),
		zap.String("timeframe", timeframe))

	// Query last 24 hours
	endTime := time.Now().Unix() * 1000
	startTime := time.Now().Add(-24*time.Hour).Unix() * 1000

	candles, err := repo.Select(ctx, symbol, exchange, timeframe, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to query candles: %w", err)
	}

	log.Info("Query results", zap.Int("count", len(candles)))

	// Display first and last few candles
	if len(candles) > 0 {
		log.Info("First candle",
			zap.Time("timestamp", time.Unix(candles[0].Timestamp/1000, 0)),
			zap.Float64("open", candles[0].Open),
			zap.Float64("high", candles[0].High),
			zap.Float64("low", candles[0].Low),
			zap.Float64("close", candles[0].Close),
			zap.Float64("volume", candles[0].Volume))

		if len(candles) > 1 {
			last := candles[len(candles)-1]
			log.Info("Last candle",
				zap.Time("timestamp", time.Unix(last.Timestamp/1000, 0)),
				zap.Float64("open", last.Open),
				zap.Float64("high", last.High),
				zap.Float64("low", last.Low),
				zap.Float64("close", last.Close),
				zap.Float64("volume", last.Volume))
		}
	}

	return nil
}

// deleteCandles removes old candle data (safety limited to >1 hour old).
func deleteCandles(ctx context.Context, repo candle.Repository, symbol, exchange, timeframe string, log *zap.Logger) error {
	log.Info("Checking candles for deletion",
		zap.String("symbol", symbol),
		zap.String("exchange", exchange),
		zap.String("timeframe", timeframe))

	// For safety, only query test data older than 1 hour
	endTime := time.Now().Add(-1*time.Hour).Unix() * 1000
	startTime := int64(0) // From beginning of time

	// First, query to see what we would delete
	candles, err := repo.Select(ctx, symbol, exchange, timeframe, startTime, endTime)
	if err != nil {
		return fmt.Errorf("failed to query candles before deletion: %w", err)
	}

	log.Warn("Candles that would be deleted",
		zap.Int("count", len(candles)),
		zap.String("symbol", symbol),
		zap.String("exchange", exchange),
		zap.String("timeframe", timeframe))

	// Note: The current repository doesn't have a Delete method, so we'll just log what we would delete
	log.Info("Delete operation simulation completed (no actual deletion performed)")
	log.Info("To implement actual deletion, add a Delete method to the candle repository")

	return nil
}

// getTimeframeInterval converts timeframe string to time.Duration.
func getTimeframeInterval(timeframe string) time.Duration {
	switch timeframe {
	case "1m":
		return time.Minute
	case "3m":
		return 3 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return time.Hour
	case "2h":
		return 2 * time.Hour
	case "4h":
		return 4 * time.Hour
	case "6h":
		return 6 * time.Hour
	case "12h":
		return 12 * time.Hour
	case "1d", "D":
		return 24 * time.Hour
	case "1w", "W":
		return 7 * 24 * time.Hour
	default:
		return time.Minute
	}
}

// getBasePriceForSymbol returns a realistic base price for different symbols.
func getBasePriceForSymbol(symbol string) float64 {
	switch symbol {
	case "BTCUSDT":
		return 50000.0
	case "ETHUSDT":
		return 3000.0
	case "BNBUSDT":
		return 400.0
	case "ADAUSDT":
		return 1.0
	case "DOGEUSDT":
		return 0.25
	default:
		return 100.0 // Default price
	}
}

// generateRealisticOHLCV creates realistic OHLCV data with proper relationships.
func generateRealisticOHLCV(basePrice float64, timestamp int64) candle_model.OHLCV {
	// Generate realistic price movement (±2% volatility)
	volatility := 0.002
	priceChange := (rand.Float64() - 0.5) * volatility
	open := basePrice * (1 + priceChange)

	// Generate high and low with realistic spreads
	highOffset := rand.Float64() * volatility * basePrice * 0.5
	lowOffset := rand.Float64() * volatility * basePrice * 0.5
	high := open + highOffset
	low := open - lowOffset

	// Ensure high >= open and low <= open
	if high < open {
		high = open
	}
	if low > open {
		low = open
	}

	// Close should be between high and low
	close := low + rand.Float64()*(high-low)

	// Generate realistic volume
	baseVolume := 100.0
	volumeVariation := 1.0 + (rand.Float64()-0.5)*0.8 // ±40% variation
	volume := baseVolume * volumeVariation

	return candle_model.OHLCV{
		Timestamp: timestamp,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    volume,
	}
}
