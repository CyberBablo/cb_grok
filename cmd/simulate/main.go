package main

import (
	"cb_grok/config"
	"cb_grok/internal/exchange"
	"cb_grok/internal/exchange/bybit"
	"cb_grok/internal/utils"
	"cb_grok/internal/utils/logger"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	Version = "dev"
)

const (
	wsAddr = "localhost:8080"
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

		// Lifecycle hooks
		fx.Invoke(registerLifecycleHooks),

		// FX settings
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

func runSimulation() error {
	var (
		symbol      string
		timeframe   string
		tradingDays int
	)
	flag.StringVar(&symbol, "symbol", "", "Symbol (f.e BNB/USDT)")
	flag.StringVar(&timeframe, "timeframe", "", "Timeframe (f.e 1h)")
	flag.IntVar(&tradingDays, "trading-days", 0, "Trading days")
	flag.Parse()

	log.Info("starting ws server", zap.String("symbol", symbol), zap.String("timeframe", timeframe), zap.Int("trading-days", tradingDays))

	if err := runServer(symbol, timeframe, tradingDays); err != nil {
		zap.L().Error(fmt.Sprintf("run ws server error: %s", err.Error()), zap.String("symbol", symbol), zap.String("timeframe", timeframe), zap.Int("trading-days", tradingDays))
		return err
	}

	return nil
}

// Server function - Entry point 1
func runServer(symbol string, timeframe string, tradingDays int) error {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ex, err := bybit.NewBybit("", "", "live")
	if err != nil {
		return err
	}

	timeframeSec := utils.TimeframeToMilliseconds(timeframe) / 1000
	candlesPerDay := (24 * 60 * 60) / int(timeframeSec)

	totalCandles := tradingDays * candlesPerDay

	candles, err := ex.FetchSpotOHLCV(symbol, exchange.Timeframe(timeframe), totalCandles)
	if err != nil {
		return err
	}

	log.Info("simulate: ohlcv data", zap.Int("length", len(candles)))

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			zap.S().Errorf("simulation server: upgrade error: %s", err.Error())
			return
		}
		defer conn.Close()

		bar := progressbar.Default(int64(len(candles)), "candles")
		for i := range candles {
			d, err := json.Marshal(candles[i])
			if err != nil {
				continue
			}
			err = conn.WriteMessage(websocket.TextMessage, d)
			if err != nil {
				zap.S().Errorf("simulation server: write message error: %s", err.Error())
				return
			}
			bar.Add(1)

			time.Sleep(10 * time.Millisecond)
		}

		// Отправляем сообщение о закрытии перед завершением
		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Simulation completed")
		err = conn.WriteMessage(websocket.CloseMessage, closeMsg)
		if err != nil {
			zap.S().Errorf("simulation server: failed to send close message: %s", err.Error())
		}
		log.Info("Simulation completed, connection closed")
	})

	zap.S().Infof("Starting WebSocket server on %s", wsAddr)
	err = http.ListenAndServe(wsAddr, nil)
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
	shutdowner fx.Shutdowner,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting simulation",
				zap.String("version", cfg.App.Version),
				zap.String("environment", cfg.App.Environment),
			)

			exitCode := 0
			go func() {
				err := runSimulation()
				if err != nil {
					log.Error("Failed to run optimize", zap.Error(err))
					exitCode = 1
				}

				_ = shutdowner.Shutdown(fx.ExitCode(exitCode))
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("Stopping simulation")

			log.Info("Simulation stopped")
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
