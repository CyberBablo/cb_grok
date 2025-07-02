package main

import (
	"cb_grok/config"
	"cb_grok/internal/candle"
	candleRepository "cb_grok/internal/candle/repository"
	"cb_grok/internal/exchange"
	"cb_grok/internal/exchange/bybit"
	metrics_usecase "cb_grok/internal/metrics/usecase"
	"cb_grok/internal/model"
	"cb_grok/internal/order"
	orderRepository "cb_grok/internal/order/repository"
	orderUsecase "cb_grok/internal/order/usecase"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/trader"
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

		fx.Provide(func(repo order.Repository, log *zap.Logger) order.Usecase {
			return orderUsecase.New(repo, log)
		}),

		fx.Provide(func(db postgres.Postgres) candle.Repository {
			return candleRepository.New(db)
		}),

		// Modules
		telegram.Module,
		trader.Module,

		// Lifecycle hooks
		fx.Invoke(
			fx.Annotate(
				registerLifecycleHooks,
				fx.ParamTags(``, ``, ``, ``, ``, ``, ``, `name:"metrics"`, ``),
			),
		),

		// FX settings
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

func runTrade(log *zap.Logger, tg *telegram.TelegramService, cfg *config.Config, orderUC order.Usecase, candleRepo candle.Repository, metricsDB postgres.Postgres) error {
	var (
		tradingMode string
	)
	flag.StringVar(&tradingMode, "trading_mode", "simulation", "Trading mode (simulation, demo, live). Default: simulation")
	flag.Parse()

	mod, err := model.Load(cfg.DemoTrading.Model)
	if err != nil {
		log.Error("Failed to load model params", zap.Error(err))
		return fmt.Errorf("error to load model: %w", err)
	}

	var exch exchange.Exchange

	switch tradingMode {
	case "demo":
		exch, err = bybit.NewBybit(cfg.Bybit.APIKey, cfg.Bybit.APISecret, exchange.TradingModeDemo)
	//case "live":
	//exch, err = bybit.NewBybit(cfg.Bybit.APIKey, cfg.Bybit.APISecret, exchange.TradingModeLive)
	default:
		exch = exchange.NewMockExchange()
	}

	if err != nil {
		log.Error("Failed to initialize exchange", zap.Error(err), zap.String("trading_mode", tradingMode))
		return fmt.Errorf("failed to initialize exchange: %w", err)
	}

	trade := trader.NewTrader(log, tg, orderUC, candleRepo)

	trade.Setup(trader.TraderParams{
		Model:          mod,
		Exchange:       exch,
		Strategy:       strategy.NewLinearBiasStrategy(),
		Settings:       nil, // default
		InitialCapital: 10000,
	})
	symbol := "BTCUSDT"
	metricsCollector := metrics_usecase.NewDBMetricsCollector(trade.GetState(), metricsDB, symbol, log)
	trade.SetMetricsCollector(metricsCollector)

	/*
		Available timeframes:

		1 3 5 15 30 (min)
		60 120 240 360 720 (min)
		D (day)
		W (week)
		M (month)
	*/
	if tradingMode == "demo" {
		err = trade.Run(trader.ModeLiveDemo, "60")
	} else {
		err = trade.RunSimulation(trader.ModeSimulation)
	}
	if err != nil {
		return err
	}

	return nil
}

// registerLifecycleHooks registers lifecycle hooks for the application
func registerLifecycleHooks(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	cfg *config.Config,
	orderUC order.Usecase,
	candleRepo candle.Repository,
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
				err := runTrade(log, tg, cfg, orderUC, candleRepo, metricsDB)
				if err != nil {
					log.Error("Failed to run optimize", zap.Error(err))
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
