package main

import (
	"cb_grok/config"
	"cb_grok/internal/backtest"
	"cb_grok/internal/exchange"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/trading_model"
	"cb_grok/internal/utils/logger"
	"flag"
	"fmt"
	"github.com/c-bata/goptuna"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"math"
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
	var (
		validationSetDays int
		symbol            string
	)
	flag.IntVar(&validationSetDays, "val-set-days", 0, "Number of days for validation set")
	flag.StringVar(&symbol, "symbol", "", "Symbol (f.e BNB/USDT)")
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

	candles, err := ex.FetchOHLCV(symbol, "30m", 5000)
	if err != nil {
		log.Error("optimize: fetch OHLCV", zap.Error(err))
		return err
	}

	log.Info("optimize: OHLCV data", zap.Int("length", len(candles)))

	candlesPerDay := 48
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
		maShortPeriod, err := trial.SuggestInt("ma_short_period", 5, 60)
		if err != nil {
			return 0, err
		}
		maLongPeriod, err := trial.SuggestInt("ma_long_period", 20, 200)
		if err != nil {
			return 0, err
		}
		rsiPeriod, err := trial.SuggestInt("rsi_period", 5, 40)
		if err != nil {
			return 0, err
		}
		atrPeriod, err := trial.SuggestInt("atr_period", 5, 40)
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
		emaShortPeriod, err := trial.SuggestInt("ema_short_period", 10, 100)
		if err != nil {
			return 0, err
		}
		emaLongPeriod, err := trial.SuggestInt("ema_long_period", 50, 400)
		if err != nil {
			return 0, err
		}
		atrThreshold, err := trial.SuggestFloat("atr_threshold", 0.0, 2.0)
		if err != nil {
			return 0, err
		}
		macdShortPeriod, err := trial.SuggestInt("macd_short_period", 5, 15)
		if err != nil {
			return 0, err
		}
		macdLongPeriod, err := trial.SuggestInt("macd_long_period", 20, 50)
		if err != nil {
			return 0, err
		}
		macdSignalPeriod, err := trial.SuggestInt("macd_signal_period", 5, 15)
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

		bollingerPeriod, err := trial.SuggestInt("bollinger_period", 10, 50)
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
			Pair:                symbol,
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

	err = study.Optimize(objective, 7550)
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
		Pair:                symbol,
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

	filename := trading_model.SaveModel(symbol, bestStrategyParams)

	orderCount := len(orders)
	result := fmt.Sprintf(
		"Оптимизация завершена.\nСимвол: %s\nКоличество сделок: %d\nКомбинированный Sharpe Ratio: %.2f\nВалидационный Sharpe Ratio: %.2f\nИтоговый капитал: %.2f\nМаксимальная просадка: %.2f%%\nWin Rate: %.2f%%\nМодель сохранена в %s",
		symbol, orderCount, combinedSharpeRatio, valSharpe, capital, valMaxDD, valWinRate, filename)

	log.Info("optimization completed",
		zap.Float64("combined_sharpe_ratio", combinedSharpeRatio),
		zap.Float64("validation_sharpe_ratio", valSharpe),
		zap.Float64("validation_max_drawdown", valMaxDD),
		zap.Float64("validation_win_rate", valWinRate),
		zap.String("filename", filename))
	tg.SendMessage(result)

	return nil
}
