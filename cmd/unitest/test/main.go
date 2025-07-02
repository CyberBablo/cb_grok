// Package test provides functionality to test Bybit exchange operations
// including wallet balance checks and order status queries.
package test

import (
	"context"
	"fmt"
	"os"

	bybit "github.com/bybit-exchange/bybit.go.api"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"cb_grok/config"
	"cb_grok/internal/exchange"
	bybit_uc "cb_grok/internal/exchange/bybit"
	"cb_grok/pkg/logger"
)

// CMD runs the test command with FX dependency injection.
func CMD() {
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

		// Provide Bybit exchange
		fx.Provide(func(cfg *config.Config) (exchange.Exchange, error) {
			return bybit_uc.NewBybit(cfg.Bybit.APIKey, cfg.Bybit.APISecret, "demo")
		}),

		// Lifecycle management
		fx.Invoke(runTestOperation),

		// FX logger
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

// runTestOperation is the main FX lifecycle function that executes test operations.
func runTestOperation(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	bybitApp exchange.Exchange,
	cfg *config.Config,
	shutdowner fx.Shutdowner,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting Bybit exchange test with FX")

			go func() {
				defer shutdowner.Shutdown()

				// Test exchange name
				log.Info("Testing exchange connection", zap.String("exchange", bybitApp.Name()))
				fmt.Printf("Connected to exchange: %s\n", bybitApp.Name())

				// Test fetching OHLCV data (this is available on the exchange interface)
				log.Info("Testing OHLCV data fetch...")
				candles, err := bybitApp.FetchSpotOHLCV("BTCUSDT", exchange.Timeframe("1h"), 10)
				if err != nil {
					log.Error("Failed to fetch OHLCV data", zap.Error(err))
				} else {
					log.Info("OHLCV Data", zap.Int("candles_count", len(candles)))
					fmt.Printf("Fetched %d candles for BTCUSDT\n", len(candles))
					if len(candles) > 0 {
						fmt.Printf("Last candle: Open=%.2f, High=%.2f, Low=%.2f, Close=%.2f, Volume=%.2f\n",
							candles[len(candles)-1].Open,
							candles[len(candles)-1].High,
							candles[len(candles)-1].Low,
							candles[len(candles)-1].Close,
							candles[len(candles)-1].Volume)
					}
				}

				// Test raw Bybit API call
				log.Info("Testing raw Bybit API call...")
				if err := testBybitAPIRequest(cfg, log); err != nil {
					log.Error("Raw API test failed", zap.Error(err))
				} else {
					log.Info("Raw API test completed successfully")
				}

				log.Info("✅ All tests completed successfully")
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Test operation completed")
			return nil
		},
	})
}

// testBybitAPIRequest tests raw Bybit API functionality.
func testBybitAPIRequest(cfg *config.Config, log *zap.Logger) error {
	var clientOptions []bybit.ClientOption
	clientOptions = append(clientOptions, bybit.WithBaseURL(bybit.DEMO_ENV), bybit.WithDebug(false))

	client := bybit.NewBybitHttpClient(cfg.Bybit.APIKey, cfg.Bybit.APISecret, clientOptions...)
	params := map[string]interface{}{"accountType": "UNIFIED"}

	response, err := client.NewUtaBybitServiceWithParams(params).GetAccountWallet(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get account wallet: %w", err)
	}

	log.Info("Account Wallet Response", zap.String("response", bybit.PrettyPrint(response)))
	fmt.Println("Raw API Response:", bybit.PrettyPrint(response))
	return nil
}
