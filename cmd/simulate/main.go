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
	fx.New(
		logger.Module,
		config.Module,
		exchange.Module,
		strategy.Module,
		backtest.Module,
		telegram.Module,
		fx.Invoke(runSimulation),
	).Run()
}

func runSimulation(
	log *zap.Logger,
	ex exchange.Exchange,
	strat strategy.Strategy,
	bt backtest.Backtest,
	tg *telegram.TelegramService,
	cfg config.Config,
) {
	var modelPath string
	var days int
	flag.StringVar(&modelPath, "model", "", "Path to model JSON file")
	flag.IntVar(&days, "days", 30, "Number of days to simulate")
	flag.Parse()

	if modelPath == "" {
		log.Error("Model path is required")
		return
	}

	modelFile, err := os.Open(modelPath)
	if err != nil {
		log.Error("Failed to open model file", zap.Error(err))
		return
	}
	defer modelFile.Close()

	var modelParams map[string]interface{}
	decoder := json.NewDecoder(modelFile)
	if err := decoder.Decode(&modelParams); err != nil {
		log.Error("Failed to decode model parameters", zap.Error(err))
		return
	}

	params := strategy.StrategyParams{
		MAShortPeriod:        modelParams["ma_short_period"].(int),
		MALongPeriod:         modelParams["ma_long_period"].(int),
		RSIPeriod:            modelParams["rsi_period"].(int),
		ATRPeriod:            modelParams["atr_period"].(int),
		BuyRSIThreshold:      modelParams["buy_rsi"].(float64),
		SellRSIThreshold:     modelParams["sell_rsi"].(float64),
		StopLossMultiplier:   modelParams["stop_loss_multiplier"].(float64),
		TakeProfitMultiplier: modelParams["take_profit_multiplier"].(float64),
		EMAShortPeriod:       modelParams["ema_short_period"].(int),
		EMALongPeriod:        modelParams["ema_long_period"].(int),
		UseTrendFilter:       modelParams["use_trend_filter"].(string) == "true",
		UseRSIFilter:         modelParams["use_rsi_filter"].(string) == "true",
		UseADXFilter:         modelParams["use_adx_filter"].(string) == "true",
		ATRThreshold:         modelParams["atr_threshold"].(float64),
		ADXPeriod:            modelParams["adx_period"].(int),
		ADXThreshold:         modelParams["adx_threshold"].(float64),
		MACDShortPeriod:      modelParams["macd_short_period"].(int),
		MACDLongPeriod:       modelParams["macd_long_period"].(int),
		MACDSignalPeriod:     modelParams["macd_signal_period"].(int),
		
		MAWeight:             modelParams["ma_weight"].(float64),
		MACDWeight:           modelParams["macd_weight"].(float64),
		RSIWeight:            modelParams["rsi_weight"].(float64),
		ADXWeight:            modelParams["adx_weight"].(float64),
		TrendWeight:          modelParams["trend_weight"].(float64),
		
		BuyThreshold:         modelParams["buy_threshold"].(float64),
		SellThreshold:        modelParams["sell_threshold"].(float64),
	}

	pair := "BNB/USDT"
	candles, err := ex.FetchOHLCV(pair, "15m", days * 96) // 96 candles per day for 15m timeframe
	if err != nil {
		log.Error("Failed to fetch OHLCV data", zap.Error(err))
		return
	}

	sharpeRatio, orders, finalCapital, maxDrawdown, winRate, err := bt.Run(candles, params)
	if err != nil {
		log.Error("Failed to run backtest", zap.Error(err))
		return
	}

	log.Info("Simulation results",
		zap.Float64("sharpe_ratio", sharpeRatio),
		zap.Int("order_count", len(orders)),
		zap.Float64("final_capital", finalCapital),
		zap.Float64("max_drawdown", maxDrawdown),
		zap.Float64("win_rate", winRate),
	)

	result := fmt.Sprintf(
		"Результат симуляции:\nВалютная пара: %s\nКоличество сделок: %d\nSharpe Ratio: %.2f\nИтоговый капитал: %.2f\nМаксимальная просадка: %.2f%%\nWin Rate: %.2f%%",
		pair, len(orders), sharpeRatio, finalCapital, maxDrawdown, winRate)
	tg.SendMessage(result)
}
