package main

import (
	"cb_grok/config"
	"cb_grok/internal/candle"
	candleRepository "cb_grok/internal/candle/repository"
	"cb_grok/internal/launcher"
	"cb_grok/internal/order"
	orderRepository "cb_grok/internal/order/repository"
	orderUsecase "cb_grok/internal/order/usecase"
	stage_model "cb_grok/internal/stage/model"
	"cb_grok/internal/strategy"
	strategyRepository "cb_grok/internal/strategy/repository"
	"cb_grok/internal/symbol"
	symbolRepository "cb_grok/internal/symbol/repository"
	"cb_grok/internal/telegram"
	trader "cb_grok/internal/trader"
	traderRepository "cb_grok/internal/trader/repository"
	"cb_grok/internal/utils/logger"
	"cb_grok/pkg/postgres"
	"context"
	"flag"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

const (
	stopLossMultiplier    = 5
	takeProfitsMultiplier = 5
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
		fx.Provide(
			fx.Annotate(
				func(cfg *config.Config) (postgres.Postgres, error) {
					return postgres.InitPsqlDB(&postgres.Conn{
						Host:     cfg.PostgresMetrics.Host,
						Port:     cfg.PostgresMetrics.Port,
						User:     cfg.PostgresMetrics.User,
						Password: cfg.PostgresMetrics.Password,
						DBName:   cfg.PostgresMetrics.DBName,
						SSLMode:  cfg.PostgresMetrics.SSLMode,
						PgDriver: cfg.PostgresMetrics.PgDriver,
					})
				},
				fx.ResultTags(`name:"metrics"`),
			),
		),

		fx.Provide(func(db postgres.Postgres) order.Repository {
			return orderRepository.New(db)
		}),

		fx.Provide(func(db postgres.Postgres) strategy.Repository { return strategyRepository.New(db) }),

		fx.Provide(func(db postgres.Postgres) trader.Repository { return traderRepository.New(db) }),

		fx.Provide(func(db postgres.Postgres) symbol.Repository { return symbolRepository.New(db) }),

		fx.Provide(func(repo order.Repository, log *zap.Logger) order.Order {
			return orderUsecase.New(repo, log)
		}),

		fx.Provide(func(db postgres.Postgres) candle.Repository {
			return candleRepository.New(db)
		}),

		// Modules
		telegram.Module,
		trader.Module,

		// Lifecycle hooks

		/*
			lifecycle fx.Lifecycle,
			log *zap.Logger,
			cfg *config.Config,
			orderUC order.Order,
			candleRepo candle.Repository,
			strategyRepo strategy.Repository,
			traderRepo trader.Repository,
			tg *telegram.TelegramService,
			db postgres.Postgres,
			metricsDB postgres.Postgres,
			shutdowner fx.Shutdowner,
		*/

		fx.Invoke(
			fx.Annotate(
				registerLifecycleHooks,
				fx.ParamTags(``, ``, ``, ``, ``, ``, ``, ``, ``, ``, `name:"metrics"`, ``),
			),
		),

		// FX settings
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

func runTrade(
	log *zap.Logger,
	tg *telegram.TelegramService,
	cfg *config.Config,
	orderUC order.Order,
	candleRepo candle.Repository,
	metricsDB postgres.Postgres,
	strategyRepo strategy.Repository,
	traderRepo trader.Repository,
	symbolRepo symbol.Repository,
) error {
	var (
		tradingMode string
	)
	flag.StringVar(&tradingMode, "trading_mode", "simulation", "Trading mode (simulation, demo, live). Default: simulation")
	flag.Parse()

	var runStage stage_model.StageStatus
	switch tradingMode {
	case "simulation":
		runStage = stage_model.StageSimulation
	case "demo":
		runStage = stage_model.StageDemo
	default:
		runStage = stage_model.StageProd
	}
	return launcher.Launch(runStage, log, tg, cfg, orderUC, candleRepo, metricsDB, strategyRepo, traderRepo, symbolRepo)
}

// registerLifecycleHooks registers lifecycle hooks for the application
func registerLifecycleHooks(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	cfg *config.Config,
	orderUC order.Order,
	candleRepo candle.Repository,
	strategyRepo strategy.Repository,
	traderRepo trader.Repository,
	symbolRepo symbol.Repository,
	tg *telegram.TelegramService,
	db postgres.Postgres,
	metricsDB postgres.Postgres,
	shutdowner fx.Shutdowner,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting backtest",
				zap.String("version", cfg.App.Version),
				zap.String("environment", cfg.App.Environment),
			)

			exitCode := 0
			go func() {
				err := runTrade(log, tg, cfg, orderUC, candleRepo, metricsDB, strategyRepo, traderRepo, symbolRepo)
				if err != nil {
					log.Error("Failed to run trade", zap.Error(err))
					exitCode = 1
				}

				_ = shutdowner.Shutdown(fx.ExitCode(exitCode))
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping trader")

			log.Info("Trader stopped")
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
