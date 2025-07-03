package main

import (
	"cb_grok/config"
	"cb_grok/internal/candle"
	candleRepository "cb_grok/internal/candle/repository"
	"cb_grok/internal/utils/logger"
	"cb_grok/pkg/models"
	"cb_grok/pkg/postgres"
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(cfg.Postgres)

	app := fx.New(
		// Configuration
		fx.Provide(func() *config.Config { return cfg }),

		// Logger
		fx.Provide(func(cfg *config.Config) (*zap.Logger, error) {
			return logger.NewZapLogger(logger.ZapConfig{
				Level:       cfg.Logger.Level,
				Development: cfg.Logger.Development,
				Encoding:    cfg.Logger.Encoding,
				OutputPaths: cfg.Logger.OutputPaths,
			})
		}),

		// Postgres
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

		// Candle repository
		fx.Provide(func(db postgres.Postgres) candle.Repository {
			return candleRepository.New(db)
		}),

		// Lifecycle hooks
		fx.Invoke(runCandleTest),

		// FX settings
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

func runCandleTest(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	candleRepo candle.Repository,
	shutdowner fx.Shutdowner,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var (
				operation string
				symbol    string
				exchange  string
				timeframe string
				count     int
			)

			flag.StringVar(&operation, "op", "generate", "Operation: generate, query, delete")
			flag.StringVar(&symbol, "symbol", "BTCUSDT", "Trading symbol")
			flag.StringVar(&exchange, "exchange", "bybit", "Exchange name")
			flag.StringVar(&timeframe, "timeframe", "1m", "Timeframe")
			flag.IntVar(&count, "count", 100, "Number of candles to generate")
			flag.Parse()

			go func() {
				defer shutdowner.Shutdown()

				switch operation {
				case "generate":
					err := generateAndSaveCandles(ctx, log, candleRepo, symbol, exchange, timeframe, count)
					if err != nil {
						log.Error("Failed to generate candles", zap.Error(err))
					}

				case "query":
					err := queryCandles(ctx, log, candleRepo, symbol, exchange, timeframe)
					if err != nil {
						log.Error("Failed to query candles", zap.Error(err))
					}

				case "delete":
					err := deleteCandles(ctx, log, candleRepo, symbol, exchange, timeframe)
					if err != nil {
						log.Error("Failed to delete candles", zap.Error(err))
					}

				default:
					log.Error("Unknown operation", zap.String("operation", operation))
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Candle test completed")
			return nil
		},
	})
}

func generateAndSaveCandles(ctx context.Context, log *zap.Logger, repo candle.Repository, symbol, exchange, timeframe string, count int) error {
	log.Info("Generating and saving candles",
		zap.String("symbol", symbol),
		zap.String("exchange", exchange),
		zap.String("timeframe", timeframe),
		zap.Int("count", count))

	// Start from current time and go backwards
	endTime := time.Now()
	interval := getTimeframeInterval(timeframe)

	// Generate realistic OHLCV data
	basePrice := 50000.0 // Base price for BTC
	if symbol != "BTCUSDT" {
		basePrice = 100.0 // Default for other symbols
	}

	candles := make([]models.OHLCV, count)

	for i := 0; i < count; i++ {
		// Calculate timestamp for this candle
		candleTime := endTime.Add(-time.Duration(count-i-1) * interval)
		timestamp := candleTime.Unix() * 1000 // Convert to milliseconds

		// Generate realistic OHLCV values
		volatility := 0.002 // 0.2% volatility
		open := basePrice * (1 + (rand.Float64()-0.5)*volatility)

		// Generate high and low
		highOffset := rand.Float64() * volatility * basePrice
		lowOffset := rand.Float64() * volatility * basePrice
		high := open + highOffset
		low := open - lowOffset

		// Close should be between high and low
		close := low + rand.Float64()*(high-low)

		// Volume in base currency
		volume := 100 + rand.Float64()*900 // Between 100 and 1000

		candles[i] = models.OHLCV{
			Timestamp: timestamp,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    volume,
		}

		// Update base price for next candle
		basePrice = close
	}

	// Save candles one by one
	successCount := 0
	for i, candle := range candles {
		err := repo.Create(ctx, symbol, exchange, timeframe, candle)
		if err != nil {
			log.Error("Failed to save candle",
				zap.Error(err),
				zap.Int("index", i),
				zap.Int64("timestamp", candle.Timestamp))
			continue
		}
		successCount++

		if (i+1)%10 == 0 {
			log.Info("Progress", zap.Int("saved", i+1), zap.Int("total", count))
		}
	}

	log.Info("Candle generation completed",
		zap.Int("requested", count),
		zap.Int("saved", successCount))

	return nil
}

func queryCandles(ctx context.Context, log *zap.Logger, repo candle.Repository, symbol, exchange, timeframe string) error {
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

func deleteCandles(ctx context.Context, log *zap.Logger, repo candle.Repository, symbol, exchange, timeframe string) error {
	log.Info("Deleting candles",
		zap.String("symbol", symbol),
		zap.String("exchange", exchange),
		zap.String("timeframe", timeframe))

	// For safety, only delete test data older than 1 hour
	endTime := time.Now().Add(-1*time.Hour).Unix() * 1000
	startTime := 0 // From beginning of time

	// First, query to see what we're about to delete
	candles, err := repo.Select(ctx, symbol, exchange, timeframe, int64(startTime), endTime)
	if err != nil {
		return fmt.Errorf("failed to query candles before deletion: %w", err)
	}

	log.Warn("About to delete candles",
		zap.Int("count", len(candles)),
		zap.String("symbol", symbol),
		zap.String("exchange", exchange),
		zap.String("timeframe", timeframe))

	// Note: The current repository doesn't have a Delete method, so we'll just log what we would delete
	log.Info("Delete operation would remove the above candles (not implemented in repository)")

	return nil
}

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
