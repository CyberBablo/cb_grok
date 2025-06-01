package main

import (
	"cb_grok/config"
	"cb_grok/internal/backtest"
	"cb_grok/internal/exchange"
	"cb_grok/internal/exchange/bybit"
	"cb_grok/internal/model"
	"cb_grok/internal/telegram"
	"cb_grok/internal/trader"
	"cb_grok/internal/utils"
	"cb_grok/internal/utils/logger"
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
	"time"
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

		// Modules
		backtest.Module,
		telegram.Module,
		trader.Module,

		// Lifecycle hooks
		fx.Invoke(registerLifecycleHooks),

		// FX settings
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

func runBacktest(cfg *config.Config, backtest backtest.Backtest, tg *telegram.TelegramService) error {

	var (
		modelFilename string
		setDays       int
		timeframe     string
	)

	flag.StringVar(&timeframe, "timeframe", "", "Timeframe (f.e 1h)")
	flag.IntVar(&setDays, "set-days", 0, "Number of days for trading set")
	flag.StringVar(&modelFilename, "model", "", "Model filename")
	flag.Parse()

	ex, err := bybit.NewBybit(cfg.Bybit.APIKey, cfg.Bybit.APISecret, exchange.TradingModeLive)
	if err != nil {
		log.Error("backtest: initialize exchange", zap.Error(err))
		return err
	}

	mod, err := model.Load(modelFilename)
	if err != nil {
		log.Error("Failed to load model params", zap.Error(err))
		return fmt.Errorf("error to load model: %w", err)
	}

	timeframeSec := utils.TimeframeToMilliseconds(timeframe) / 1000
	candlesPerDay := (24 * 60 * 60) / int(timeframeSec)

	candlesTotal := setDays * candlesPerDay

	candles, err := ex.FetchSpotOHLCV(mod.Symbol, exchange.Timeframe(timeframe), candlesTotal)
	if err != nil {
		zap.L().Error("backtest: fetch ohlcv", zap.Error(err))
		return err
	}

	result, err := backtest.Run(candles, mod)
	if err != nil {
		zap.L().Error("backtest: run backtest", zap.Error(err))
		return err
	}

	zap.L().Info("backtest completed")

	msg := fmt.Sprintf(
		"Результат бектеста:\n\nМодель: %s\nСимвол: %s\nTimeframe: %s\nКол-во свечей: %d\nКол-во дней на валидации: %d\nКоличество сделок: %d\nSharpe Ratio: %.2f\nИтоговый капитал: %.2f\nМаксимальная просадка: %.2f%%\nWin Rate: %.2f%%",
		modelFilename, mod.Symbol, timeframe, len(result.TradeState.GetOHLCV()), setDays, len(result.Orders), result.SharpeRatio, result.FinalCapital, result.MaxDrawdown, result.WinRate)

	time.Sleep(1000 * time.Millisecond)
	chartBuff, err := result.TradeState.GenerateCharts()
	if err != nil {
		zap.L().Error("report: generate charts", zap.Error(err))
	}

	err = tg.SendFile(chartBuff, "html", msg)
	if err != nil {
		zap.L().Error("report: send to telegram", zap.Error(err))
	}

	zap.L().Info("report sent to Telegram")

	return nil

}

// registerLifecycleHooks registers lifecycle hooks for the application
func registerLifecycleHooks(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	cfg *config.Config,
	tg *telegram.TelegramService,
	backtest backtest.Backtest,
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
				err := runBacktest(cfg, backtest, tg)
				if err != nil {
					log.Error("Failed to run backtest", zap.Error(err))
					exitCode = 1
				}

				_ = shutdowner.Shutdown(fx.ExitCode(exitCode))
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping backtest")

			log.Info("Backtest stopped")
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
