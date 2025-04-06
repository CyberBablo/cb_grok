package indicators

import (
	"cb_grok/pkg/models"
	"github.com/cinar/indicator"
	"math"
)

func CalculateSMA(candles []models.OHLCV, period int) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	sma := indicator.Sma(period, closes)
	return sma
}

func CalculateEMA(candles []models.OHLCV, period int) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	ema := indicator.Ema(period, closes)
	return ema
}

func CalculateRSI(candles []models.OHLCV, period int) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	_, rsi := indicator.RsiPeriod(period, closes) // Используем RsiPeriod, игнорируем rs
	return rsi
}

func CalculateATR(candles []models.OHLCV, period int) []float64 {
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}
	_, atr := indicator.Atr(period, highs, lows, closes) // Игнорируем tr
	return atr
}

func CalculateMACD(candles []models.OHLCV, shortPeriod, longPeriod, signalPeriod int) ([]float64, []float64) {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	macd, signal := indicator.Macd(closes) // Возвращаем только macd и signal
	return macd, signal
}

func CalculateBollingerBands(candles []models.OHLCV, period int, stdDevMultiplier float64) ([]float64, []float64, []float64) {
	if len(candles) < period {
		return nil, nil, nil // Недостаточно данных
	}

	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	// Рассчитываем SMA
	middle := CalculateSMA(candles, period) // Используем существующую функцию CalculateSMA
	if middle == nil {
		return nil, nil, nil
	}

	// Рассчитываем стандартное отклонение
	stdDev := make([]float64, len(middle))
	for i := range middle {
		start := i
		end := i + period
		if end > len(closes) {
			end = len(closes)
		}
		slice := closes[start:end]
		var sumSquares float64
		for _, price := range slice {
			diff := price - middle[i]
			sumSquares += diff * diff
		}
		stdDev[i] = math.Sqrt(sumSquares / float64(len(slice)))
	}

	// Рассчитываем верхнюю и нижнюю полосы
	upper := make([]float64, len(middle))
	lower := make([]float64, len(middle))
	for i := range middle {
		upper[i] = middle[i] + stdDevMultiplier*stdDev[i]
		lower[i] = middle[i] - stdDevMultiplier*stdDev[i]
	}

	return upper, middle, lower
}
