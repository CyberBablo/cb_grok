package main

import (
	"cb_grok/config"
	"cb_grok/internal/backtest"
	"cb_grok/internal/exchange"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/utils"
	"cb_grok/internal/utils/logger"
	"encoding/json"
	"flag"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

func main() {
	fx.New(
		logger.Module,
		config.Module,
		exchange.Module,
		backtest.Module,
		telegram.Module,
		fx.Invoke(runOptimization),
	).Run()
}

func ReadStrategyParams(filePath string) (strategy.StrategyParams, error) {
	var params strategy.StrategyParams

	// Read the file using os.ReadFile
	data, err := os.ReadFile(filePath)
	if err != nil {
		return params, err
	}

	// Parse the JSON data into the struct
	err = json.Unmarshal(data, &params)
	if err != nil {
		return params, err
	}

	return params, nil
}

func runOptimization(
	log *zap.Logger,
	bt backtest.Backtest,
	tg *telegram.TelegramService,
	cfg config.Config,
) error {
	var validationSetDays int
	var symbol string
	var timeDelta string
	var trials int
	var dataSetCandles int
	var modelPath string

	flag.IntVar(&dataSetCandles, "dataset-candles", 1000, "Number of dataset Candles")
	flag.StringVar(&symbol, "symbol", "BTC/USDT", "Number of days for validation set")
	flag.StringVar(&timeDelta, "time-delta", "", "Time delta for validation set")
	flag.StringVar(&modelPath, "model-path", "", "Path to model json file")
	flag.Parse()

	bestStrategyParams, err := ReadStrategyParams(modelPath)

	ex, err := exchange.NewBinance(false, cfg.Binance.ApuPublic, cfg.Binance.ApiSecret, cfg.Binance.ProxyUrl)
	if err != nil {
		log.Error("optimize: initialize Binance exchange", zap.Error(err))
		return err
	}

	secondsOfDelta := utils.TimeframeToMilliseconds(timeDelta) / 1000
	candlesPerDay := (24 * 60 * 60) / int(secondsOfDelta)
	candles, err := ex.FetchOHLCV(symbol, timeDelta, dataSetCandles)
	if err != nil {
		log.Error("optimize: fetch OHLCV", zap.Error(err))
		return err
	}

	valSharpe, orders, capital, valMaxDD, valWinRate, err := bt.RunIterativeApply(candles, bestStrategyParams)
	if err != nil {
		log.Error("optimize: final validation backtest", zap.Error(err))
		return err
	}

	fmt.Println("ORDER HISTORY")
	for _, order := range orders {
		log.Info(fmt.Sprintf("Order: %v", order))
	}

	dir := "lib/best_models"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Error("optimize: create dir", zap.Error(err))
		return err
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(dir, fmt.Sprintf("model_%s.json", timestamp))
	file, err := os.Create(filename)
	if err != nil {
		log.Error("optimize: create new model file", zap.Error(err))
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(bestStrategyParams); err != nil {
		log.Error("optimize: model file encoder", zap.Error(err))
		return err
	}
	orderCount := len(orders)

	resultIterative := fmt.Sprintf(
		"РЕЗУЛЬТАТ ИТЕРАТИВНОГО БЕКТЕСТА\nПара: %s\nКоличество trials: %d\nКоличество дней на валидации: %d\nТаймдельта: %s\nКоличество свечей в сутках: %d\nКоличество сделок: %d\nВалидационный Sharpe Ratio: %.2f\nИтоговый капитал: %.2f\nМаксимальная просадка: %.2f%%\nWin Rate: %.2f%%\nМодель сохранена в %s",
		symbol, trials, validationSetDays, timeDelta, candlesPerDay, orderCount, valSharpe, capital, valMaxDD, valWinRate, filename)

	log.Info("optimization completed",
		zap.Float64("validation_sharpe_ratio", valSharpe),
		zap.Float64("validation_max_drawdown", valMaxDD),
		zap.Float64("validation_win_rate", valWinRate),
		zap.String("filename", filename))

	tg.SendMessage(resultIterative)

	time.Sleep(5000 * time.Millisecond)
	os.Exit(0)
	return nil
}
