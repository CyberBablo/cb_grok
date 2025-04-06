package strategy

import (
	"cb_grok/internal/indicators"
	"cb_grok/pkg/models"
)

type MovingAverageStrategy struct{}

func NewMovingAverageStrategy() Strategy {
	return &MovingAverageStrategy{}
}

func (s *MovingAverageStrategy) Apply(candles []models.OHLCV, params StrategyParams) []models.AppliedOHLCV {
	if len(candles) < max(params.MALongPeriod, params.EMALongPeriod, params.MACDLongPeriod) {
		return nil
	}

	shortMA := indicators.CalculateSMA(candles, params.MAShortPeriod)
	longMA := indicators.CalculateSMA(candles, params.MALongPeriod)
	rsi := indicators.CalculateRSI(candles, params.RSIPeriod)
	atr := indicators.CalculateATR(candles, params.ATRPeriod)
	emaShort := indicators.CalculateEMA(candles, params.EMAShortPeriod)
	emaLong := indicators.CalculateEMA(candles, params.EMALongPeriod)
	macd, macdSignal := indicators.CalculateMACD(candles, params.MACDShortPeriod, params.MACDLongPeriod, params.MACDSignalPeriod)

	trend := make([]bool, len(candles))
	volatility := make([]bool, len(candles))
	for i := range candles {
		trend[i] = emaShort[i] > emaLong[i]
		volatility[i] = atr[i] > params.ATRThreshold
	}

	var appliedCandles []models.AppliedOHLCV
	for i := range candles {
		appliedCandles = append(appliedCandles, models.AppliedOHLCV{
			OHLCV:      candles[i],
			ATR:        atr[i],
			RSI:        rsi[i],
			ShortMA:    shortMA[i],
			LongMA:     longMA[i],
			ShortEMA:   emaShort[i],
			LongEMA:    emaLong[i],
			Trend:      trend[i],
			Volatility: volatility[i],
			MACD:       macd[i],
			MACDSignal: macdSignal[i],
		})
	}

	for i := 1; i < len(appliedCandles); i++ {
		signals := Signals{EMASignal: 0, RSISignal: 0, MACDSignal: 0, TrendSignal: 0}

		if shortMA[i] > longMA[i] {
			signals.EMASignal = 1
		} else if shortMA[i] < longMA[i] {
			signals.EMASignal = -1
		}
		if macd[i] > macdSignal[i] {
			signals.MACDSignal = 1
		} else if macd[i] < macdSignal[i] {
			signals.MACDSignal = -1
		}

		if trend[i] && volatility[i] {
			signals.TrendSignal = 1
		} else if !trend[i] && volatility[i] {
			signals.TrendSignal = -1
		}

		if rsi[i] < params.BuyRSIThreshold {
			signals.RSISignal = 1
		} else if rsi[i] > params.SellRSIThreshold {
			signals.RSISignal = -1
		}

		totalWeight := params.RSIWeight + params.MACDWeight + params.TrendWeight + params.EMAWeight
		params.RSIWeight = params.RSIWeight / totalWeight
		params.TrendWeight = params.TrendWeight / totalWeight
		params.MACDWeight = params.MACDWeight / totalWeight
		params.EMAWeight = params.EMAWeight / totalWeight

		signal :=
			float64(signals.RSISignal)*params.RSIWeight +
				float64(signals.MACDSignal)*params.MACDWeight +
				float64(signals.TrendSignal)*params.TrendWeight +
				float64(signals.EMASignal)*params.EMAWeight

		signal = Tanh(signal)

		if signal > params.BuySignalThreshold {
			//fmt.Println("SIGNAL", signal, params.BuySignalThreshold, params.SellSignalThreshold)
			appliedCandles[i].Signal = 1
		} else if signal < params.SellSignalThreshold {
			//fmt.Println("SIGNAL", signal, params.BuySignalThreshold, params.SellSignalThreshold)

			appliedCandles[i].Signal = -1
		} else {
			appliedCandles[i].Signal = 0
		}
		//fmt.Println("SIGNALS", signals, signal)
		//fmt.Println("WEIGHTS", params.RSIWeight, params.MACDWeight, params.TrendWeight, params.EMAWeight)
		//fmt.Println("FINAL SIGNAL", appliedCandles[i].Signal)

		if i > 0 {
			appliedCandles[i].Position = appliedCandles[i].Signal - appliedCandles[i-1].Signal
		}
	}

	return appliedCandles
}

func max(a, b, c int) int {
	if a > b {
		if a > c {
			return a
		}
		return c
	}
	if b > c {
		return b
	}
	return c
}
