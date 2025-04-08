package strategy

import (
	"cb_grok/internal/indicators"
	"cb_grok/pkg/models"
	"math"
)

type LinearBiasStrategy struct{}

func NewLinearBiasStrategy() Strategy {
	return &LinearBiasStrategy{}
}

func (s *LinearBiasStrategy) Apply(candles []models.OHLCV, params StrategyParams) []models.AppliedOHLCV {
	// Обновлено условие: добавлен StochasticKPeriod для проверки минимальной длины данных
	if len(candles) < max(params.MALongPeriod, params.EMALongPeriod, params.MACDLongPeriod, params.BollingerPeriod, params.StochasticKPeriod) {
		return nil
	}

	// Рассчитываем ATR для адаптивных периодов
	atr := indicators.CalculateATR(candles, params.ATRPeriod)

	// Рассчитываем индикаторы с использованием статических периодов
	shortMA := indicators.CalculateSMA(candles, params.MAShortPeriod)
	longMA := indicators.CalculateSMA(candles, params.MALongPeriod)
	rsi := indicators.CalculateRSI(candles, params.RSIPeriod)
	emaShort := indicators.CalculateEMA(candles, params.EMAShortPeriod)
	emaLong := indicators.CalculateEMA(candles, params.EMALongPeriod)
	macd, macdSignal := indicators.CalculateMACD(candles, params.MACDShortPeriod, params.MACDLongPeriod, params.MACDSignalPeriod)
	upperBB, _, lowerBB := indicators.CalculateBollingerBands(candles, params.BollingerPeriod, params.BollingerStdDev)
	stochasticK, stochasticD := indicators.CalculateStochasticOscillator(candles, params.StochasticKPeriod, params.StochasticDPeriod)

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
			UpperBB:    upperBB[i],
			LowerBB:    lowerBB[i],
			// Добавлены поля для Stochastic Oscillator
			StochasticK: stochasticK[i],
			StochasticD: stochasticD[i],
		})
	}

	for i := 1; i < len(appliedCandles); i++ {
		// Добавлено поле StochasticSignal в инициализацию структуры Signals
		signals := Signals{EMASignal: 0, RSISignal: 0, MACDSignal: 0, TrendSignal: 0, BBSignal: 0, StochasticSignal: 0}

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

		if candles[i].Close < appliedCandles[i].LowerBB {
			signals.BBSignal = 1
		} else if candles[i].Close > appliedCandles[i].UpperBB {
			signals.BBSignal = -1
		}

		// Добавлена логика сигналов для Stochastic Oscillator
		if stochasticK[i] < 20 && stochasticK[i] > stochasticD[i] {
			signals.StochasticSignal = 1 // Покупка при пересечении снизу вверх в зоне перепроданности
		} else if stochasticK[i] > 80 && stochasticK[i] < stochasticD[i] {
			signals.StochasticSignal = -1 // Продажа при пересечении сверху вниз в зоне перекупленности
		}

		// Обновлен расчет общего веса с учетом StochasticWeight
		totalWeight := params.RSIWeight + params.MACDWeight + params.TrendWeight + params.EMAWeight + params.BBWeight + params.StochasticWeight
		if totalWeight == 0 {
			totalWeight = 1
		}
		params.RSIWeight = params.RSIWeight / totalWeight
		params.TrendWeight = params.TrendWeight / totalWeight
		params.MACDWeight = params.MACDWeight / totalWeight
		params.EMAWeight = params.EMAWeight / totalWeight
		params.BBWeight = params.BBWeight / totalWeight
		params.StochasticWeight = params.StochasticWeight / totalWeight

		// Добавлен вклад StochasticSignal в итоговый сигнал
		signal := float64(signals.RSISignal)*params.RSIWeight +
			float64(signals.MACDSignal)*params.MACDWeight +
			float64(signals.TrendSignal)*params.TrendWeight +
			float64(signals.EMASignal)*params.EMAWeight +
			float64(signals.BBSignal)*params.BBWeight +
			float64(signals.StochasticSignal)*params.StochasticWeight

		signal = math.Tanh(signal)

		if signal > params.BuySignalThreshold {
			appliedCandles[i].Signal = 1
		} else if signal < params.SellSignalThreshold {
			appliedCandles[i].Signal = -1
		} else {
			appliedCandles[i].Signal = 0
		}

		if i > 0 {
			appliedCandles[i].Position = appliedCandles[i].Signal - appliedCandles[i-1].Signal
		}
	}

	return appliedCandles
}

func calculateATRMean(atr []float64) float64 {
	if len(atr) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range atr {
		sum += v
	}
	return sum / float64(len(atr))
}

func adjustPeriod(basePeriod int, atr []float64, atrMean float64) int {
	if atrMean == 0 || len(atr) == 0 {
		return basePeriod
	}
	// Адаптация периода: увеличиваем при высокой волатильности
	adjustment := 1 + (atr[len(atr)-1] / atrMean)
	adjusted := int(math.Round(float64(basePeriod) * adjustment))
	if adjusted < 1 {
		adjusted = 1
	}
	return adjusted
}
