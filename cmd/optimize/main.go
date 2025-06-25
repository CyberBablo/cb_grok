package main

import (
	"cb_grok/config"
	"cb_grok/internal/backtest"
	"cb_grok/internal/candle"
	candleRepository "cb_grok/internal/candle/repository"
	"cb_grok/internal/optimize"
	"cb_grok/internal/order"
	orderRepository "cb_grok/internal/order/repository"
	orderUsecase "cb_grok/internal/order/usecase"
	"cb_grok/internal/telegram"
	"cb_grok/internal/utils/logger"
	"cb_grok/pkg/postgres"
	"context"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

var (
	Version = "dev"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

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
				Host:     cfg.Postgres.Host,
				Port:     cfg.Postgres.Port,
				User:     cfg.Postgres.User,
				Password: cfg.Postgres.Password,
				DBName:   cfg.Postgres.DBName,
				SSLMode:  cfg.Postgres.SSLMode,
				PgDriver: cfg.Postgres.PgDriver,
			})
		}),

		fx.Provide(func(db postgres.Postgres) order.Repository {
			return orderRepository.New(db)
		}),

		fx.Provide(func(repo order.Repository, log *zap.Logger) order.Usecase {
			return orderUsecase.New(repo, log)
		}),

		fx.Provide(func(db postgres.Postgres) candle.Repository {
			return candleRepository.New(db)
		}),

		// Modules
		optimize.Module,
		telegram.Module,
		backtest.Module,

		// Lifecycle hooks
		fx.Invoke(registerLifecycleHooks),

		// FX settings
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

func runOptimize(cfg *config.Config, opt optimize.Optimize) error {
	log.Info("Starting optimize",
		zap.String("version", cfg.App.Version),
		zap.String("environment", cfg.App.Environment),
	)

	var (
		trainSetDays int
		valSetDays   int
		symbol       string
		timeframe    string
		trials       int
		workers      int
	)

	flag.StringVar(&symbol, "symbol", "", "Symbol (f.e BNB/USDT)")
	flag.StringVar(&timeframe, "timeframe", "", "Timeframe (f.e 1h)")
	flag.IntVar(&trials, "trials", 100, "Number of trials")
	flag.IntVar(&trainSetDays, "train-set-days", 0, "Number of days for training set")
	flag.IntVar(&valSetDays, "val-set-days", 0, "Number of days for validation set")
	flag.IntVar(&workers, "workers", 2, "Number of parallel workers")

	flag.Parse()

	return opt.Run(optimize.RunOptimizeParams{
		Symbol:       symbol,
		Timeframe:    timeframe,
		TrainSetDays: trainSetDays,
		ValSetDays:   valSetDays,
		Trials:       trials,
		Workers:      workers,
	})
}

// registerLifecycleHooks registers lifecycle hooks for the application
func registerLifecycleHooks(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	cfg *config.Config,
	opt optimize.Optimize,
	shutdowner fx.Shutdowner,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting optimize",
				zap.String("version", cfg.App.Version),
				zap.String("environment", cfg.App.Environment),
			)

			exitCode := 0
			go func() {
				err := runOptimize(cfg, opt)
				if err != nil {
					log.Error("Failed to run optimize", zap.Error(err))
					exitCode = 1
				}

				_ = shutdowner.Shutdown(fx.ExitCode(exitCode))
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping optimize")

			log.Info("Optimize stopped")
			return nil
		},
	})

	// Handle OS signals
	go handleSignals(log)
}

// handleSignals handles OS signals for graceful shutdown
func handleSignals(log *zap.Logger) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Info("Received signal", zap.String("signal", sig.String()))

	// fx will handle graceful shutdown automatically
}
