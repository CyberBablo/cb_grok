// Package place_order provides functionality to test order placement
// and account management operations on Bybit exchange.
package place_order

import (
	"context"
	"fmt"
	"os"

	bybit "github.com/bybit-exchange/bybit.go.api"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"cb_grok/config"
	"cb_grok/pkg/logger"
)

// CMD runs the place_order command with FX dependency injection.
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

		// Provide Bybit HTTP client
		fx.Provide(func(cfg *config.Config) *bybit.Client {
			var clientOptions []bybit.ClientOption
			clientOptions = append(clientOptions, bybit.WithBaseURL(bybit.DEMO_ENV), bybit.WithDebug(false))
			return bybit.NewBybitHttpClient(cfg.Bybit.APIKey, cfg.Bybit.APISecret, clientOptions...)
		}),

		// Lifecycle management
		fx.Invoke(runPlaceOrderOperation),

		// FX logger
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

// runPlaceOrderOperation is the main FX lifecycle function that executes place order operations.
func runPlaceOrderOperation(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	client *bybit.Client,
	shutdowner fx.Shutdowner,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting place order test with FX")

			go func() {
				defer shutdowner.Shutdown()

				// Test account wallet functionality
				log.Info("Testing account wallet query...")
				params := map[string]interface{}{"accountType": "UNIFIED"}

				response, err := client.NewUtaBybitServiceWithParams(params).GetAccountWallet(ctx)
				if err != nil {
					log.Error("Failed to get account wallet", zap.Error(err))
					return
				}

				log.Info("Account Wallet Response", zap.String("response", bybit.PrettyPrint(response)))
				fmt.Println("Account Wallet:", bybit.PrettyPrint(response))

				// Note: Actual order placement is commented out for safety
				// Uncomment and modify the following code to test actual order placement
				/*
					log.Info("Testing order placement (commented out for safety)...")
					orderParams := map[string]interface{}{
						"category":    "spot",
						"symbol":      "BTCUSDT",
						"side":        "Buy",
						"orderType":   "Market",
						"qty":         "0.001",
					}
					orderResponse, err := client.NewTradeService().PlaceOrder(ctx, orderParams)
					if err != nil {
						log.Error("Failed to place order", zap.Error(err))
						return
					}
					log.Info("Order Response", zap.String("response", bybit.PrettyPrint(orderResponse)))
				*/

				log.Info("✅ Place order test completed successfully")
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Place order operation completed")
			return nil
		},
	})
}
