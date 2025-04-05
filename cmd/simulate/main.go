package main

import (
	"cb_grok/config"
	"cb_grok/internal/backtest"
	"cb_grok/internal/exchange"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/utils/logger"
	"encoding/json"
	"flag"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
)

func main() {
	modelPath := flag.String("model", "", "Path to the model JSON file")
	flag.Parse()

	if *modelPath == "" {
		fmt.Println("Please provide a model path with -model flag")
		os.Exit(1)
	}

	fx.New(
		logger.Module,
		config.Module,
		exchange.Module,
		strategy.Module,
		backtest.Module,
		telegram.Module,
		fx.Invoke(func(log *zap.Logger, ex exchange.Exchange, strat strategy.Strategy, bt backtest.Backtest, tg *telegram.TelegramService, cfg config.Config) {
			runSimulation(log, ex, strat, bt, tg, cfg, *modelPath)
		}),
	).Run()
}

func runSimulation(
	log *zap.Logger,
	ex exchange.Exchange,
	strat strategy.Strategy,
	bt backtest.Backtest,
	tg *telegram.TelegramService,
	cfg config.Config,
	modelPath string,
) {
	modelFile, err := os.ReadFile(modelPath)
	if err != nil {
		log.Error("Failed to read model file", zap.Error(err))
		return
	}

	var params strategy.StrategyParams
	if err := json.Unmarshal(modelFile, &params); err != nil {
		log.Error("Failed to parse model parameters", zap.Error(err))
		return
	}

	candles, err := ex.FetchOHLCV("BNB/USDT", "1h", 10000)
	if err != nil {
		log.Error("Failed to fetch OHLCV data", zap.Error(err))
		return
	}

	log.Info("Running simulation with model parameters", 
		zap.String("model", modelPath),
		zap.Int("candles", len(candles)))

	sharpeRatio, orders, finalCapital, maxDrawdown, winRate, err := bt.Run(candles, params)
	if err != nil {
		log.Error("Backtest error", zap.Error(err))
		return
	}

	log.Info("Simulation results", 
		zap.Float64("sharpe_ratio", sharpeRatio),
		zap.Float64("final_capital", finalCapital),
		zap.Float64("max_drawdown", maxDrawdown),
		zap.Float64("win_rate", winRate),
		zap.Int("total_trades", len(orders)/2))

	message := fmt.Sprintf("Simulation Results:\n- Sharpe Ratio: %.2f\n- Final Capital: %.2f\n- Max Drawdown: %.2f%%\n- Win Rate: %.2f%%\n- Total Trades: %d",
		sharpeRatio, finalCapital, maxDrawdown, winRate*100, len(orders)/2)
	tg.SendMessage(message)
}
