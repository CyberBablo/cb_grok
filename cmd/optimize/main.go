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
	"github.com/samber/lo"
	"go.uber.org/fx"
	"go.uber.org/zap"
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
		maShortPeriod, err := trial.SuggestInt("ma_short_period", 5, 30)
		if err != nil {
			return 0, err
		}
		maLongPeriod, err := trial.SuggestInt("ma_long_period", 20, 100)
		if err != nil {
			return 0, err
		}
		rsiPeriod, err := trial.SuggestInt("rsi_period", 5, 20)
		if err != nil {
			return 0, err
		}
		atrPeriod, err := trial.SuggestInt("atr_period", 5, 20)
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
		stopLossMultiplier, err := trial.SuggestFloat("stop_loss_multiplier", 0.5, 3.0)
		if err != nil {
			return 0, err
		}
		takeProfitMultiplier, err := trial.SuggestFloat("take_profit_multiplier", 1.0, 5.0)
		if err != nil {
			return 0, err
		}
		emaShortPeriod, err := trial.SuggestInt("ema_short_period", 10, 50)
		if err != nil {
			return 0, err
		}
		emaLongPeriod, err := trial.SuggestInt("ema_long_period", 50, 200)
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
		atrThreshold, err := trial.SuggestFloat("atr_threshold", 0.0, 2.0)
		if err != nil {
			return 0, err
		}
		adxPeriod, err := trial.SuggestInt("adx_period", 10, 30)
		if err != nil {
			return 0, err
		}
		adxThreshold, err := trial.SuggestFloat("adx_threshold", 20, 40)
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
			ADXPeriod:            adxPeriod,
			ADXThreshold:         adxThreshold,
			MACDShortPeriod:      macdShortPeriod,
			MACDLongPeriod:       macdLongPeriod,
			MACDSignalPeriod:     macdSignalPeriod,
		}

		trainSharpe, _, _, trainMaxDD, trainWinRate, err := bt.Run(trainCandles, params)
		if err != nil {
			return 0, err
		}

		valSharpe, _, _, valMaxDD, valWinRate, err := bt.Run(validationCandles, params)
		if err != nil {
			log.Warn("Failed to run backtest on validation set", zap.Error(err))
			return trainSharpe, nil
		}

		combinedSharpe := (trainSharpe + valSharpe) / 2
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
		ADXPeriod:            bestParams["adx_period"].(int),
		ADXThreshold:         bestParams["adx_threshold"].(float64),
		MACDShortPeriod:      bestParams["macd_short_period"].(int),
		MACDLongPeriod:       bestParams["macd_long_period"].(int),
		MACDSignalPeriod:     bestParams["macd_signal_period"].(int),
	}

	valSharpe, orders, capital, valMaxDD, valWinRate, err := bt.Run(validationCandles, bestStrategyParams)
	if err != nil {
		log.Error("optimize: final validation backtest", zap.Error(err))
		return err
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
