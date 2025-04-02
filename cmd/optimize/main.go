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
	"github.com/c-bata/goptuna"
	"github.com/samber/lo"
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

func runOptimization(
	log *zap.Logger,
	bt backtest.Backtest,
	tg *telegram.TelegramService,
) error {
	var validationSetDays int
	flag.IntVar(&validationSetDays, "val-set-days", 0, "Number of days for validation set")
	flag.Parse()

	if validationSetDays <= 0 {
		log.Error("Validation days must be greater than 0")
		return fmt.Errorf("validation days must be greater than 0")
	}

	ex, err := exchange.NewBybit(false, "", "")
	if err != nil {
		log.Error("optimize: initialize Bybit exchange", zap.Error(err))
		return err
	}

	candles, err := ex.FetchOHLCV("BNB/USDT", "1h", 3000)
	if err != nil {
		log.Error("optimize: fetch OHLCV", zap.Error(err))
		return err
	}

	log.Info("optimize: OHLCV data", zap.Int("length", len(candles)))

	candlesPerDay := 24
	validationCandlesCount := validationSetDays * candlesPerDay

	if validationCandlesCount >= len(candles) {
		log.Error("Validation set is larger than the available data")
		return fmt.Errorf("validation set is larger than the available data")
	}

	validationCandles := candles[len(candles)-validationCandlesCount:]
	trainCandles := candles[:len(candles)-validationCandlesCount]

	log.Info("optimize: datasets prepared",
		zap.Int("train_candles", len(trainCandles)),
		zap.Int("validation_candles", len(validationCandles)))

	study, err := goptuna.CreateStudy(
		"strategy_1",
		goptuna.StudyOptionDirection(goptuna.StudyDirectionMaximize),
	)
	if err != nil {
		log.Error("optimize: create study", zap.Error(err))
		return err
	}

	objective := func(trial goptuna.Trial) (float64, error) {
		maShortPeriod, err := trial.SuggestInt("ma_short_period", 5, 15)
		if err != nil {
			return 0, err
		}
		maLongPeriod, err := trial.SuggestInt("ma_long_period", 20, 50)
		if err != nil {
			return 0, err
		}
		rsiPeriod, err := trial.SuggestInt("rsi_period", 8, 16)
		if err != nil {
			return 0, err
		}
		atrPeriod, err := trial.SuggestInt("atr_period", 8, 16)
		if err != nil {
			return 0, err
		}
		buyRsiThreshold, err := trial.SuggestFloat("buy_rsi", 15, 35)
		if err != nil {
			return 0, err
		}
		sellRsiThreshold, err := trial.SuggestFloat("sell_rsi", 65, 85)
		if err != nil {
			return 0, err
		}
		stopLossMultiplier, err := trial.SuggestFloat("stop_loss_multiplier", 0.8, 2.0)
		if err != nil {
			return 0, err
		}
		takeProfitMultiplier, err := trial.SuggestFloat("take_profit_multiplier", 1.5, 3.5)
		if err != nil {
			return 0, err
		}
		emaShortPeriod, err := trial.SuggestInt("ema_short_period", 15, 40)
		if err != nil {
			return 0, err
		}
		emaLongPeriod, err := trial.SuggestInt("ema_long_period", 80, 150)
		if err != nil {
			return 0, err
		}
		useTrendFilter, err := trial.SuggestCategorical("use_trend_filter", []string{"true", "false"})
		if err != nil {
			return 0, err
		}
		useRsiFilter, err := trial.SuggestCategorical("use_rsi_filter", []string{"true", "false"})
		if err != nil {
			return 0, err
		}
		atrThreshold, err := trial.SuggestFloat("atr_threshold", 0.0, 1.0)
		if err != nil {
			return 0, err
		}

		params := strategy.StrategyParams{
			MAShortPeriod:        maShortPeriod,
			MALongPeriod:         maLongPeriod,
			RSIPeriod:            rsiPeriod,
			ATRPeriod:            atrPeriod,
			BuyRSIThreshold:      buyRsiThreshold,
			SellRSIThreshold:     sellRsiThreshold,
			StopLossMultiplier:   stopLossMultiplier,
			TakeProfitMultiplier: takeProfitMultiplier,
			EMAShortPeriod:       emaShortPeriod,
			EMALongPeriod:        emaLongPeriod,
			UseTrendFilter:       lo.If(useTrendFilter == "true", true).Else(false),
			UseRSIFilter:         lo.If(useRsiFilter == "true", true).Else(false),
			ATRThreshold:         atrThreshold,
		}

		trainSharpe, _, err := bt.Run(trainCandles, params)
		if err != nil {
			return 0, err
		}

		valSharpe, _, err := bt.Run(validationCandles, params)
		if err != nil {
			log.Warn("Failed to run backtest on validation set", zap.Error(err))
		}

		combinedSharpe := (trainSharpe + valSharpe) / 2

		return combinedSharpe, nil
	}

	err = study.Optimize(objective, 1000)
	if err != nil {
		log.Error("optimize: study optimize", zap.Error(err))
		return err
	}

	bestParams, err := study.GetBestParams()
	if err != nil {
		log.Error("optimize: get best params", zap.Error(err))
		return err
	}
	combinedSharpeRatio, err := study.GetBestValue()
	if err != nil {
		log.Error("optimize: get best value", zap.Error(err))
		return err
	}

	bestStrategyParams := strategy.StrategyParams{
		MAShortPeriod:        bestParams["ma_short_period"].(int),
		MALongPeriod:         bestParams["ma_long_period"].(int),
		RSIPeriod:            bestParams["rsi_period"].(int),
		ATRPeriod:            bestParams["atr_period"].(int),
		BuyRSIThreshold:      bestParams["buy_rsi"].(float64),
		SellRSIThreshold:     bestParams["sell_rsi"].(float64),
		StopLossMultiplier:   bestParams["stop_loss_multiplier"].(float64),
		TakeProfitMultiplier: bestParams["take_profit_multiplier"].(float64),
		EMAShortPeriod:       bestParams["ema_short_period"].(int),
		EMALongPeriod:        bestParams["ema_long_period"].(int),
		UseTrendFilter:       bestParams["use_trend_filter"].(string) == "true",
		UseRSIFilter:         bestParams["use_rsi_filter"].(string) == "true",
		ATRThreshold:         bestParams["atr_threshold"].(float64),
	}

	valSharpe, _, err := bt.Run(validationCandles, bestStrategyParams)
	if err != nil {
		log.Error("optimize: final validation backtest", zap.Error(err))
		return err
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
	if err := encoder.Encode(bestParams); err != nil {
		log.Error("optimize: model file encoder", zap.Error(err))
		return err
	}

	result := fmt.Sprintf(
		"Оптимизация завершена.\nКомбинированный Sharpe Ratio: %.2f\nВалидационный Sharpe Ratio: %.2f\nМодель сохранена в %s",
		combinedSharpeRatio, valSharpe, filename)

	log.Info("optimization completed",
		zap.Float64("combined_sharpe_ratio", combinedSharpeRatio),
		zap.Float64("validation_sharpe_ratio", valSharpe),
		zap.String("filename", filename))
	tg.SendMessage(result)

	return nil
}
