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
	flag.IntVar(&validationSetDays, "val-set-days", 0, "Number of days for validation set")
	flag.Parse()

	if validationSetDays <= 0 {
		log.Error("Validation days must be greater than 0")
		return fmt.Errorf("validation days must be greater than 0")
	}

	ex, err := exchange.NewBinance(false, cfg.Binance.ApiPublic, cfg.Binance.ApiSecret, cfg.Binance.ProxyUrl)
	if err != nil {
		log.Error("optimize: initialize Binance exchange", zap.Error(err))
		return err
	}
	pair := "BNB/USDT"
	candles, err := ex.FetchOHLCV(pair, "15m", 6000)
	if err != nil {
		log.Error("optimize: fetch OHLCV", zap.Error(err))
		return err
	}

	log.Info("optimize: OHLCV data", zap.Int("length", len(candles)))

	candlesPerDay := 96
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
	//os.Exit(0)
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
		useAdxFilter, err := trial.SuggestCategorical("use_adx_filter", []string{"true", "false"})
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
		adxThreshold, err := trial.SuggestFloat("adx_threshold", 15, 40)
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
		
		maWeight, err := trial.SuggestFloat("ma_weight", 0.1, 1.0)
		if err != nil {
			return 0, err
		}
		macdWeight, err := trial.SuggestFloat("macd_weight", 0.1, 1.0)
		if err != nil {
			return 0, err
		}
		rsiWeight, err := trial.SuggestFloat("rsi_weight", 0.1, 1.0)
		if err != nil {
			return 0, err
		}
		adxWeight, err := trial.SuggestFloat("adx_weight", 0.1, 1.0)
		if err != nil {
			return 0, err
		}
		trendWeight, err := trial.SuggestFloat("trend_weight", 0.1, 1.0)
		if err != nil {
			return 0, err
		}
		
		buyThreshold, err := trial.SuggestFloat("buy_threshold", 0.5, 2.0)
		if err != nil {
			return 0, err
		}
		sellThreshold, err := trial.SuggestFloat("sell_threshold", 0.5, 2.0)
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
			UseADXFilter:         lo.If(useAdxFilter == "true", true).Else(false),
			ATRThreshold:         atrThreshold,
			ADXPeriod:            adxPeriod,
			ADXThreshold:         adxThreshold,
			MACDShortPeriod:      macdShortPeriod,
			MACDLongPeriod:       macdLongPeriod,
			MACDSignalPeriod:     macdSignalPeriod,
			
			MAWeight:             maWeight,
			MACDWeight:           macdWeight,
			RSIWeight:            rsiWeight,
			ADXWeight:            adxWeight,
			TrendWeight:          trendWeight,
			
			BuyThreshold:         buyThreshold,
			SellThreshold:        sellThreshold,
		}

		trainSharpe, _, _, trainMaxDD, _, err := bt.Run(trainCandles, params)
		if err != nil {
			return 0, err
		}

		valSharpe, valOrders, _, valMaxDD, valWinRate, err := bt.Run(validationCandles, params)
		if err != nil {
			log.Warn("Failed to run backtest on validation set", zap.Error(err))
			return trainSharpe, nil
		}
		
		tradesPerMonth := float64(len(valOrders)) / float64(validationSetDays) * 30.0
		
		tradeFrequencyScore := 0.0
		if tradesPerMonth >= 15.0 && tradesPerMonth <= 20.0 {
			tradeFrequencyScore = 1.0 // Perfect trade frequency
		} else if tradesPerMonth > 20.0 && tradesPerMonth <= 30.0 {
			tradeFrequencyScore = 0.9 - ((tradesPerMonth - 20.0) / 100.0) // Small penalty for too many trades
		} else if tradesPerMonth > 30.0 && tradesPerMonth <= 50.0 {
			tradeFrequencyScore = 0.8 - ((tradesPerMonth - 30.0) / 100.0) // Medium penalty for excessive trades
		} else if tradesPerMonth > 50.0 {
			tradeFrequencyScore = 0.7 - ((tradesPerMonth - 50.0) / 500.0) // Larger penalty for too many trades
			if tradeFrequencyScore < 0.3 {
				tradeFrequencyScore = 0.3 // Floor for extremely high trade counts
			}
		} else if tradesPerMonth >= 10.0 && tradesPerMonth < 15.0 {
			tradeFrequencyScore = 0.9 - ((15.0 - tradesPerMonth) / 50.0) // Small penalty for slightly too few trades
		} else if tradesPerMonth >= 5.0 && tradesPerMonth < 10.0 {
			tradeFrequencyScore = 0.8 - ((10.0 - tradesPerMonth) / 50.0) // Medium penalty for too few trades
		} else {
			tradeFrequencyScore = 0.5 * (tradesPerMonth / 5.0) // Severe penalty for almost no trades
		}
		
		winRateScore := 0.0
		if valWinRate >= 75.0 && valWinRate <= 80.0 {
			winRateScore = 1.0 // Perfect win rate
		} else if valWinRate > 80.0 && valWinRate <= 90.0 {
			winRateScore = 0.95 - ((valWinRate - 80.0) / 200.0) // Small penalty for too high win rate
		} else if valWinRate > 90.0 {
			winRateScore = 0.9 - ((valWinRate - 90.0) / 100.0) // Medium penalty for extremely high win rate
		} else if valWinRate >= 65.0 && valWinRate < 75.0 {
			winRateScore = 0.9 - ((75.0 - valWinRate) / 100.0) // Small penalty for slightly low win rate
		} else if valWinRate >= 50.0 && valWinRate < 65.0 {
			winRateScore = 0.8 - ((65.0 - valWinRate) / 75.0) // Medium penalty for low win rate
		} else if valWinRate >= 30.0 && valWinRate < 50.0 {
			winRateScore = 0.5 - ((50.0 - valWinRate) / 100.0) // Large penalty for very low win rate
		} else {
			winRateScore = 0.3 * (valWinRate / 30.0) // Severe penalty for extremely low win rate
		}
		
		sharpeScore := (trainSharpe + valSharpe) / 2
		if sharpeScore < 0 {
			sharpeScore = 0.1 + (1.0 / (1.0 + math.Abs(sharpeScore)))
		}
		
		combinedScore := sharpeScore * tradeFrequencyScore * winRateScore
		
		log.Info("Trial result",
			zap.Int("trial", trial.ID),
			zap.Float64("combined_score", combinedScore),
			zap.Float64("trades_per_month", tradesPerMonth),
			zap.Float64("win_rate", valWinRate),
			zap.Float64("trade_frequency_score", tradeFrequencyScore),
			zap.Float64("win_rate_score", winRateScore),
			zap.Float64("sharpe_score", sharpeScore),
			zap.Float64("train_max_dd", trainMaxDD),
			zap.Float64("val_max_dd", valMaxDD),
		)

		return combinedScore, nil
	}

	err = study.Optimize(objective, 10000)
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
		UseADXFilter:         bestParams["use_adx_filter"].(string) == "true",
		ATRThreshold:         bestParams["atr_threshold"].(float64),
		ADXPeriod:            bestParams["adx_period"].(int),
		ADXThreshold:         bestParams["adx_threshold"].(float64),
		MACDShortPeriod:      bestParams["macd_short_period"].(int),
		MACDLongPeriod:       bestParams["macd_long_period"].(int),
		MACDSignalPeriod:     bestParams["macd_signal_period"].(int),
		
		MAWeight:             bestParams["ma_weight"].(float64),
		MACDWeight:           bestParams["macd_weight"].(float64),
		RSIWeight:            bestParams["rsi_weight"].(float64),
		ADXWeight:            bestParams["adx_weight"].(float64),
		TrendWeight:          bestParams["trend_weight"].(float64),
		
		BuyThreshold:         bestParams["buy_threshold"].(float64),
		SellThreshold:        bestParams["sell_threshold"].(float64),
	}

	valSharpe, orders, capital, valMaxDD, valWinRate, err := bt.Run(validationCandles, bestStrategyParams)
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
	orderCount := len(orders)
	result := fmt.Sprintf(
		"Оптимизация завершена.\nВалютная пара: %s\nКоличество сделок: %d\nКомбинированный Sharpe Ratio: %.2f\nВалидационный Sharpe Ratio: %.2f\nИтоговый капитал: %.2f\nМаксимальная просадка: %.2f%%\nWin Rate: %.2f%%\nМодель сохранена в %s",
		pair, orderCount, combinedSharpeRatio, valSharpe, capital, valMaxDD, valWinRate, filename)

	log.Info("optimization completed",
		zap.Float64("combined_sharpe_ratio", combinedSharpeRatio),
		zap.Float64("validation_sharpe_ratio", valSharpe),
		zap.Float64("validation_max_drawdown", valMaxDD),
		zap.Float64("validation_win_rate", valWinRate),
		zap.String("filename", filename))
	tg.SendMessage(result)
	return nil
}
