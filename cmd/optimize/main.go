package main

import (
	"cb_grok/config"
	"cb_grok/internal/backtest"
	"cb_grok/internal/exchange"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/utils"
	"cb_grok/internal/utils/logger"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/c-bata/goptuna"
	"github.com/c-bata/goptuna/tpe"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"math"
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
	cfg config.Config,
) error {
	var validationSetDays int
	var pair string
	var timeDelta string
	var trials int
	var dataSetCandles int
	var workers int
	flag.IntVar(&trials, "trials", 30000, "Number of trials")
	flag.IntVar(&validationSetDays, "val-set-days", 0, "Number of days for validation set")
	flag.IntVar(&dataSetCandles, "dataset-candles", 1000, "Number of dataset Candles")
	flag.IntVar(&workers, "workers", 2, "Number of parallel workers")
	flag.StringVar(&pair, "trade-pair", "BTC/USDT", "Number of days for validation set")
	flag.StringVar(&timeDelta, "time-delta", "", "Time delta for validation set")
	flag.Parse()

	if validationSetDays <= 0 {
		log.Error("Validation days must be greater than 0")
		return fmt.Errorf("validation days must be greater than 0")
	}

	ex, err := exchange.NewBinance(false, cfg.Binance.ApuPublic, cfg.Binance.ApiSecret, cfg.Binance.ProxyUrl)
	if err != nil {
		log.Error("optimize: initialize Binance exchange", zap.Error(err))
		return err
	}

	secondsOfDelta := utils.TimeframeToMilliseconds(timeDelta) / 1000
	timePeriodMultiplier := float64(60 * 60 / secondsOfDelta)
	candlesPerDay := (24 * 60 * 60) / int(secondsOfDelta)
	candles, err := ex.FetchOHLCV(pair, timeDelta, dataSetCandles)
	if err != nil {
		log.Error("optimize: fetch OHLCV", zap.Error(err))
		return err
	}
	fmt.Println("multiplier", timePeriodMultiplier)
	//os.Exit(0)
	log.Info("optimize: OHLCV data", zap.Int("length", len(candles)))

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
		goptuna.StudyOptionSampler(tpe.NewSampler()),
	)
	if err != nil {
		log.Error("optimize: create study", zap.Error(err))
		return err
	}

	objective := func(trial goptuna.Trial) (float64, error) {
		maShortPeriod, err := trial.SuggestStepInt("ma_short_period", 5, 30*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		maLongPeriod, err := trial.SuggestStepInt("ma_long_period", 20, 100*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		rsiPeriod, err := trial.SuggestStepInt("rsi_period", 5, 20*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		atrPeriod, err := trial.SuggestStepInt("atr_period", 5, 20*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		buyRsiThreshold, err := trial.SuggestFloat("buy_rsi", 10, 40)
		if err != nil {
			return 0, err
		}
		sellRsiThreshold, err := trial.SuggestFloat("sell_rsi", 60, 90)
		if err != nil {
			return 0, err
		}
		emaShortPeriod, err := trial.SuggestStepInt("ema_short_period", 10, 50*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		emaLongPeriod, err := trial.SuggestStepInt("ema_long_period", 50, 200*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		atrThreshold, err := trial.SuggestFloat("atr_threshold", 0.0, 2.0)
		if err != nil {
			return 0, err
		}
		macdShortPeriod, err := trial.SuggestStepInt("macd_short_period", 5, 15*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		macdLongPeriod, err := trial.SuggestStepInt("macd_long_period", 20, 50*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		macdSignalPeriod, err := trial.SuggestStepInt("macd_signal_period", 5, 15*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}

		buySignalThreshold, err := trial.SuggestFloat("buy_signal_threshold", 0.1, 0.9)
		if err != nil {
			return 0, err
		}

		sellSignalThreshold, err := trial.SuggestFloat("sell_signal_threshold", -0.9, -0.1)
		if err != nil {
			return 0, err
		}

		emaWeight, err := trial.SuggestFloat("ema_weight", 0, 1)
		if err != nil {
			return 0, err
		}

		trendWeight, err := trial.SuggestFloat("trend_weight", 0, 1)
		if err != nil {
			return 0, err
		}

		rsiWeight, err := trial.SuggestFloat("rsi_weight", 0, 1)
		if err != nil {
			return 0, err
		}

		macdWeight, err := trial.SuggestFloat("macd_weight", 0, 1)
		if err != nil {
			return 0, err
		}

		bollingerPeriod, err := trial.SuggestStepInt("bollinger_period", 10, 50*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		bollingerStdDev, err := trial.SuggestFloat("bollinger_std_dev", 1.0, 3.0)
		if err != nil {
			return 0, err
		}
		bbWeight, err := trial.SuggestFloat("bb_weight", 0, 1)
		if err != nil {
			return 0, err
		}

		stochasticKPeriod, err := trial.SuggestStepInt("stochastic_k_period", 5, 20*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		stochasticDPeriod, err := trial.SuggestStepInt("stochastic_d_period", 3, 10*int(timePeriodMultiplier), int(timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		stochasticWeight, err := trial.SuggestFloat("stochastic_weight", 0, 1)
		if err != nil {
			return 0, err
		}

		params := strategy.StrategyParams{
			MAShortPeriod:       maShortPeriod,
			MALongPeriod:        maLongPeriod,
			RSIPeriod:           rsiPeriod,
			ATRPeriod:           atrPeriod,
			BuyRSIThreshold:     buyRsiThreshold,
			SellRSIThreshold:    sellRsiThreshold,
			EMAShortPeriod:      emaShortPeriod,
			EMALongPeriod:       emaLongPeriod,
			ATRThreshold:        atrThreshold,
			MACDShortPeriod:     macdShortPeriod,
			MACDLongPeriod:      macdLongPeriod,
			MACDSignalPeriod:    macdSignalPeriod,
			EMAWeight:           emaWeight,
			TrendWeight:         trendWeight,
			RSIWeight:           rsiWeight,
			MACDWeight:          macdWeight,
			BuySignalThreshold:  buySignalThreshold,
			SellSignalThreshold: sellSignalThreshold,
			BollingerPeriod:     bollingerPeriod,
			BollingerStdDev:     bollingerStdDev,
			BBWeight:            bbWeight,
			StochasticDPeriod:   stochasticDPeriod,
			StochasticKPeriod:   stochasticKPeriod,
			StochasticWeight:    stochasticWeight,
			Pair:                pair,
		}

		trainSharpe, _, _, trainMaxDD, trainWinRate, err := bt.Run(trainCandles, params)
		if err != nil {
			return 0, err
		}

		valSharpe, valOrders, _, valMaxDD, valWinRate, err := bt.Run(validationCandles, params)
		if err != nil {
			log.Warn("Failed to run backtest on validation set", zap.Error(err))
			return trainSharpe, nil
		}

		//combinedSharpe := (valSharpe + trainSharpe) / 2 * math.Log(float64(len(valOrders)+1))

		combinedSharpe := valSharpe * (1 - valMaxDD/100) * math.Log(float64(len(valOrders)+1))
		//combinedSharpe = combinedSharpe * valWinRate / 100
		log.Info("Trial result",
			zap.Int("trial", trial.ID),
			zap.Float64("combined_sharpe", combinedSharpe),
			zap.Float64("train_max_dd", trainMaxDD),
			zap.Float64("train_win_rate", trainWinRate),
			zap.Float64("val_max_dd", valMaxDD),
			zap.Float64("val_win_rate", valWinRate),
		)

		return combinedSharpe, nil
	}

	eg, ctx := errgroup.WithContext(context.Background())
	study.WithContext(ctx)

	for i := 0; i < workers; i++ {
		eg.Go(func() error {
			return study.Optimize(objective, trials/workers)
		})
	}

	if err = eg.Wait(); err != nil {
		log.Error("Optimize error %v", zap.Error(err))
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
		MAShortPeriod:       bestParams["ma_short_period"].(int),
		MALongPeriod:        bestParams["ma_long_period"].(int),
		RSIPeriod:           bestParams["rsi_period"].(int),
		ATRPeriod:           bestParams["atr_period"].(int),
		BuyRSIThreshold:     bestParams["buy_rsi"].(float64),
		SellRSIThreshold:    bestParams["sell_rsi"].(float64),
		EMAShortPeriod:      bestParams["ema_short_period"].(int),
		EMALongPeriod:       bestParams["ema_long_period"].(int),
		ATRThreshold:        bestParams["atr_threshold"].(float64),
		MACDShortPeriod:     bestParams["macd_short_period"].(int),
		MACDLongPeriod:      bestParams["macd_long_period"].(int),
		MACDSignalPeriod:    bestParams["macd_signal_period"].(int),
		EMAWeight:           bestParams["ema_weight"].(float64),
		TrendWeight:         bestParams["trend_weight"].(float64),
		RSIWeight:           bestParams["rsi_weight"].(float64),
		MACDWeight:          bestParams["macd_weight"].(float64),
		BuySignalThreshold:  bestParams["buy_signal_threshold"].(float64),
		SellSignalThreshold: bestParams["sell_signal_threshold"].(float64),
		BollingerPeriod:     bestParams["bollinger_period"].(int),
		BollingerStdDev:     bestParams["bollinger_std_dev"].(float64),
		BBWeight:            bestParams["bb_weight"].(float64),
		StochasticKPeriod:   bestParams["stochastic_k_period"].(int),
		StochasticDPeriod:   bestParams["stochastic_d_period"].(int),
		StochasticWeight:    bestParams["stochastic_weight"].(float64),
		Pair:                pair,
	}

	valSharpe, orders, capital, valMaxDD, valWinRate, err := bt.Run(validationCandles, bestStrategyParams)
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
	result := fmt.Sprintf(
		"Ветка DEVX\nПара: %s\nКоличество trials: %d\nКоличество дней на валидации: %d\nТаймдельта: %s\nКоличество свечей в сутках: %d\nКоличество сделок: %d\nКомбинированный Sharpe Ratio: %.2f\nВалидационный Sharpe Ratio: %.2f\nИтоговый капитал: %.2f\nМаксимальная просадка: %.2f%%\nWin Rate: %.2f%%\nМодель сохранена в %s",
		pair, trials, validationSetDays, timeDelta, candlesPerDay, orderCount, combinedSharpeRatio, valSharpe, capital, valMaxDD, valWinRate, filename)

	log.Info("optimization completed",
		zap.Float64("combined_sharpe_ratio", combinedSharpeRatio),
		zap.Float64("validation_sharpe_ratio", valSharpe),
		zap.Float64("validation_max_drawdown", valMaxDD),
		zap.Float64("validation_win_rate", valWinRate),
		zap.String("filename", filename))
	tg.SendMessage(result)
	time.Sleep(5000 * time.Millisecond)
	os.Exit(0)
	return nil
}
