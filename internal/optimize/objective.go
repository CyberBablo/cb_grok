package optimize

import (
	"cb_grok/internal/model"
	"cb_grok/internal/strategy"
	"cb_grok/pkg/models"
	"github.com/c-bata/goptuna"
	"go.uber.org/zap"
)

type objectiveParams struct {
	symbol               string
	candles              []models.OHLCV
	setDays              int
	timePeriodMultiplier float64
}

func (o *optimize) objective(params objectiveParams) func(trial goptuna.Trial) (float64, error) {
	return func(trial goptuna.Trial) (float64, error) {
		maShortPeriod, err := trial.SuggestStepInt("ma_short_period", 5, 30*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		maLongPeriod, err := trial.SuggestStepInt("ma_long_period", 20, 100*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		rsiPeriod, err := trial.SuggestStepInt("rsi_period", 5, 20*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		atrPeriod, err := trial.SuggestStepInt("atr_period", 5, 20*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		buyRsiThreshold, err := trial.SuggestFloat("buy_rsi_threshold", 10, 40)
		if err != nil {
			return 0, err
		}
		sellRsiThreshold, err := trial.SuggestFloat("sell_rsi_threshold", 60, 90)
		if err != nil {
			return 0, err
		}
		emaShortPeriod, err := trial.SuggestStepInt("ema_short_period", 10, 50*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		emaLongPeriod, err := trial.SuggestStepInt("ema_long_period", 50, 200*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		atrThreshold, err := trial.SuggestFloat("atr_threshold", 0.0, 2.0)
		if err != nil {
			return 0, err
		}
		macdShortPeriod, err := trial.SuggestStepInt("macd_short_period", 5, 15*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		macdLongPeriod, err := trial.SuggestStepInt("macd_long_period", 20, 50*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		macdSignalPeriod, err := trial.SuggestStepInt("macd_signal_period", 5, 15*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
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

		bollingerPeriod, err := trial.SuggestStepInt("bollinger_period", 10, 50*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
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

		//bollingerPeriod := int(0)
		//bollingerStdDev := float64(0)
		//bbWeight := 0.0

		stochasticKPeriod, err := trial.SuggestStepInt("stochastic_k_period", 5, 20*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		stochasticDPeriod, err := trial.SuggestStepInt("stochastic_d_period", 3, 10*int(params.timePeriodMultiplier), int(params.timePeriodMultiplier))
		if err != nil {
			return 0, err
		}
		stochasticWeight, err := trial.SuggestFloat("stochastic_weight", 0, 1)
		if err != nil {
			return 0, err
		}

		strategyParams := strategy.StrategyParams{
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
		}

		trainBTResult, err := o.bt.Run(params.candles, &model.Model{
			Symbol:         params.symbol,
			StrategyParams: strategyParams,
		})
		if err != nil {
			return 0, err
		}

		combinedSharpe := trainBTResult.SharpeRatio * (1 - trainBTResult.MaxDrawdown/100) * min(float64(len(trainBTResult.Orders))/(float64(params.setDays*2)), 1)

		o.log.Info("Trial result",
			zap.Int("trial", trial.ID),
			zap.Float64("combined_sharpe", combinedSharpe),
			zap.Float64("train_max_dd", trainBTResult.MaxDrawdown),
			zap.Float64("train_win_rate", trainBTResult.WinRate),
			zap.Int("Orders", len(trainBTResult.Orders)),
			zap.Float64("Final capital", trainBTResult.FinalCapital),
		)

		return combinedSharpe, nil
	}
}
