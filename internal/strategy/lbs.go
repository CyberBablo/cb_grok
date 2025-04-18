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

func (s *LinearBiasStrategy) ApplyIndicators(candles []models.OHLCV, params StrategyParams) []models.AppliedOHLCV {
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
	obv := indicators.CalculateOBV(candles)                      // Новый индикатор
	vwap := indicators.CalculateVWAP(candles, params.VWAPPeriod) // Новый индикатор
	cci := indicators.CalculateCCI(candles, params.CCIPeriod)    // Новый индикатор

	trend := make([]bool, len(candles))
	volatility := make([]bool, len(candles))
	for i := range candles {
		trend[i] = emaShort[i] > emaLong[i]
		volatility[i] = atr[i] > params.ATRThreshold
	}

	var appliedCandles []models.AppliedOHLCV
	for i := range candles {
		appliedCandles = append(appliedCandles, models.AppliedOHLCV{
			OHLCV:       candles[i],
			ATR:         atr[i],
			RSI:         rsi[i],
			ShortMA:     shortMA[i],
			LongMA:      longMA[i],
			ShortEMA:    emaShort[i],
			LongEMA:     emaLong[i],
			Trend:       trend[i],
			Volatility:  volatility[i],
			MACD:        macd[i],
			MACDSignal:  macdSignal[i],
			UpperBB:     upperBB[i],
			LowerBB:     lowerBB[i],
			StochasticK: stochasticK[i],
			StochasticD: stochasticD[i],
			OBV:         obv[i],
			VWAP:        vwap[i],
			CCI:         cci[i],
		})
	}

	return appliedCandles
}

func (s *LinearBiasStrategy) ApplySignals(candles []models.AppliedOHLCV, params StrategyParams) []models.AppliedOHLCV {
	// Обновлено условие: добавлен StochasticKPeriod для проверки минимальной длины данных

	for i := 1; i < len(candles); i++ {
		// Добавлено поле StochasticSignal в инициализацию структуры Signals
		signals := Signals{EMASignal: 0, RSISignal: 0, MACDSignal: 0, TrendSignal: 0, BBSignal: 0, StochasticSignal: 0, OBVSignal: 0, VWAPSignal: 0, CCISignal: 0}

		if candles[i-1].ShortMA > candles[i-1].LongMA {
			signals.EMASignal = 1
		} else if candles[i-1].ShortMA < candles[i-1].LongMA {
			signals.EMASignal = -1
		}
		if candles[i-1].MACD > candles[i-1].MACDSignal {
			signals.MACDSignal = 1
		} else if candles[i-1].MACD < candles[i-1].MACDSignal {
			signals.MACDSignal = -1
		}

		if candles[i-1].Trend && candles[i-1].Volatility {
			signals.TrendSignal = 1
		} else if !candles[i-1].Trend && candles[i-1].Volatility {
			signals.TrendSignal = -1
		}

		if candles[i-1].RSI < params.BuyRSIThreshold {
			signals.RSISignal = 1
		} else if candles[i-1].RSI > params.SellRSIThreshold {
			signals.RSISignal = -1
		}

		if candles[i-1].Close < candles[i-1].LowerBB {
			signals.BBSignal = 1
		} else if candles[i-1].Close > candles[i-1].UpperBB {
			signals.BBSignal = -1
		}

		// Добавлена логика сигналов для Stochastic Oscillator
		if candles[i-1].StochasticK < 20 && candles[i-1].StochasticK > candles[i-1].StochasticD {
			signals.StochasticSignal = 1 // Покупка при пересечении снизу вверх в зоне перепроданности
		} else if candles[i-1].StochasticK > 80 && candles[i-1].StochasticK < candles[i-1].StochasticD {
			signals.StochasticSignal = -1 // Продажа при пересечении сверху вниз в зоне перекупленности
		}

		if i > 1 {
			// Логика для OBV
			if candles[i-1].OBV > candles[i-2].OBV {
				signals.OBVSignal = 1
			} else if candles[i-1].OBV < candles[i-2].OBV {
				signals.OBVSignal = -1
			}
		}
		// Логика для VWAP
		if candles[i-1].Close > candles[i-1].VWAP {
			signals.VWAPSignal = 1
		} else if candles[i-1].Close < candles[i-1].VWAP {
			signals.VWAPSignal = -1
		}

		// Логика для CCI
		if candles[i-1].CCI < -100 {
			signals.CCISignal = 1
		} else if candles[i-1].CCI > 100 {
			signals.CCISignal = -1
		}

		// Обновлен расчет общего веса с учетом новых весов
		totalWeight := params.RSIWeight + params.MACDWeight + params.TrendWeight + params.EMAWeight + params.BBWeight + params.StochasticWeight + params.OBVWeight + params.VWAPWeight + params.CCIWeight
		if totalWeight == 0 {
			totalWeight = 1
		}
		params.RSIWeight = params.RSIWeight / totalWeight
		params.TrendWeight = params.TrendWeight / totalWeight
		params.MACDWeight = params.MACDWeight / totalWeight
		params.EMAWeight = params.EMAWeight / totalWeight
		params.BBWeight = params.BBWeight / totalWeight
		params.StochasticWeight = params.StochasticWeight / totalWeight
		params.OBVWeight = params.OBVWeight / totalWeight
		params.VWAPWeight = params.VWAPWeight / totalWeight
		params.CCIWeight = params.CCIWeight / totalWeight

		signal := float64(signals.RSISignal)*params.RSIWeight +
			float64(signals.MACDSignal)*params.MACDWeight +
			float64(signals.TrendSignal)*params.TrendWeight +
			float64(signals.EMASignal)*params.EMAWeight +
			float64(signals.BBSignal)*params.BBWeight +
			float64(signals.StochasticSignal)*params.StochasticWeight +
			float64(signals.OBVSignal)*params.OBVWeight + // Новый вклад
			float64(signals.VWAPSignal)*params.VWAPWeight + // Новый вклад
			float64(signals.CCISignal)*params.CCIWeight // Новый вклад

		signal = math.Tanh(signal)

		if signal > params.BuySignalThreshold {
			candles[i].Signal = 1
		} else if signal < params.SellSignalThreshold {
			candles[i].Signal = -1
		} else {
			candles[i].Signal = 0
		}
	}

	return candles
}
