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
		})
	}

	return appliedCandles
}

func (s *LinearBiasStrategy) ApplySignals(candles []models.AppliedOHLCV, params StrategyParams) []models.AppliedOHLCV {
	if len(candles) < 2 {
		return candles
	}

	// Сделаем вычисление сигналов более чувствительным для генерации большего количества торговых сигналов
	for i := 1; i < len(candles); i++ {
		signals := Signals{EMASignal: 0, RSISignal: 0, MACDSignal: 0, TrendSignal: 0, BBSignal: 0, StochasticSignal: 0}

		// Увеличиваем чувствительность пересечения MA
		if candles[i-1].ShortMA <= candles[i-1].LongMA && candles[i].ShortMA > candles[i].LongMA {
			signals.EMASignal = 2 // Усиливаем сигнал пересечения снизу вверх
		} else if candles[i-1].ShortMA >= candles[i-1].LongMA && candles[i].ShortMA < candles[i].LongMA {
			signals.EMASignal = -2 // Усиливаем сигнал пересечения сверху вниз
		} else if candles[i].ShortMA > candles[i].LongMA {
			signals.EMASignal = 1 // Положительный тренд
		} else if candles[i].ShortMA < candles[i].LongMA {
			signals.EMASignal = -1 // Отрицательный тренд
		}

		// Более чувствительное пересечение MACD
		if candles[i-1].MACD <= candles[i-1].MACDSignal && candles[i].MACD > candles[i].MACDSignal {
			signals.MACDSignal = 2 // Усиливаем сигнал пересечения MACD снизу вверх
		} else if candles[i-1].MACD >= candles[i-1].MACDSignal && candles[i].MACD < candles[i].MACDSignal {
			signals.MACDSignal = -2 // Усиливаем сигнал пересечения MACD сверху вниз
		} else if candles[i].MACD > candles[i].MACDSignal {
			signals.MACDSignal = 1
		} else if candles[i].MACD < candles[i].MACDSignal {
			signals.MACDSignal = -1
		}

		// Упрощаем распознавание тренда - если есть волатильность, используем направление тренда
		if candles[i].Trend && candles[i].Volatility {
			signals.TrendSignal = 1
		} else if !candles[i].Trend && candles[i].Volatility {
			signals.TrendSignal = -1
		}

		// Более отзывчивые уровни RSI
		if candles[i].RSI < params.BuyRSIThreshold+5 { // +5 для увеличения чувствительности
			signals.RSISignal = 1
		} else if candles[i].RSI > params.SellRSIThreshold-5 { // -5 для увеличения чувствительности
			signals.RSISignal = -1
		}

		// Более чувствительные уровни Bollinger Bands
		if candles[i].Close < candles[i].LowerBB*1.01 { // Чуть менее строгое условие
			signals.BBSignal = 1
		} else if candles[i].Close > candles[i].UpperBB*0.99 { // Чуть менее строгое условие
			signals.BBSignal = -1
		}

		// Улучшенные сигналы для Stochastic Oscillator
		if candles[i].StochasticK < 25 && candles[i].StochasticK > candles[i].StochasticD {
			signals.StochasticSignal = 1 // Покупка при пересечении в зоне перепроданности
		} else if candles[i].StochasticK > 75 && candles[i].StochasticK < candles[i].StochasticD {
			signals.StochasticSignal = -1 // Продажа при пересечении в зоне перекупленности
		}

		// Правильное нормирование весов
		totalWeight := params.RSIWeight + params.MACDWeight + params.TrendWeight + params.EMAWeight + params.BBWeight + params.StochasticWeight
		if totalWeight == 0 {
			totalWeight = 1
		}

		rsiWeight := params.RSIWeight / totalWeight
		trendWeight := params.TrendWeight / totalWeight
		macdWeight := params.MACDWeight / totalWeight
		emaWeight := params.EMAWeight / totalWeight
		bbWeight := params.BBWeight / totalWeight
		stochasticWeight := params.StochasticWeight / totalWeight

		// Дополнительное усиление сигналов в определенных условиях
		// Если RSI в экстремальной зоне, усиливаем его влияние
		if candles[i].RSI < 30 || candles[i].RSI > 70 {
			rsiWeight *= 1.5
		}

		// Если цена пробила Bollinger Band, это сильный сигнал
		if signals.BBSignal != 0 {
			bbWeight *= 1.5
		}

		// Рассчитываем итоговый сигнал с учетом скорректированных весов
		signal := float64(signals.RSISignal)*rsiWeight +
			float64(signals.MACDSignal)*macdWeight +
			float64(signals.TrendSignal)*trendWeight +
			float64(signals.EMASignal)*emaWeight +
			float64(signals.BBSignal)*bbWeight +
			float64(signals.StochasticSignal)*stochasticWeight

		// Применяем функцию tanh для нормализации сигнала в диапазоне [-1, 1]
		signal = math.Tanh(signal)

		// Более чувствительные пороги для открытия позиций
		if signal > params.BuySignalThreshold*0.85 { // Снижаем порог сигнала для покупки
			candles[i].Signal = 1
		} else if signal < params.SellSignalThreshold*0.85 { // Снижаем порог сигнала для продажи
			candles[i].Signal = -1
		} else {
			candles[i].Signal = 0
		}
	}

	return candles
}
