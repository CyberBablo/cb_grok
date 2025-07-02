package indicators

import (
	candle_model "cb_grok/internal/models/candle"
	"math"
)

func CalculateBollingerBands(candles []candle_model.OHLCV, period int, stdDevMultiplier float64) ([]float64, []float64, []float64) {
	if len(candles) < period || period <= 0 || stdDevMultiplier < 0 {
		return nil, nil, nil // Недостаточно данных или некорректные параметры
	}

	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	// Рассчитываем SMA
	middle := CalculateSMA(candles, period)
	if middle == nil {
		return nil, nil, nil
	}

	// Рассчитываем стандартное отклонение
	stdDev := make([]float64, len(middle))
	for i := period - 1; i < len(closes); i++ {
		start := i - period + 1
		slice := closes[start : i+1]
		var sum, sumSquares float64
		for _, price := range slice {
			sum += price
			sumSquares += price * price
		}
		mean := sum / float64(len(slice))
		variance := (sumSquares / float64(len(slice))) - (mean * mean)
		if variance < 0 { // Избегаем отрицательной дисперсии из-за погрешностей float
			variance = 0
		}
		stdDev[i] = math.Sqrt(variance)
	}

	// Рассчитываем верхнюю и нижнюю полосы
	upper := make([]float64, len(middle))
	lower := make([]float64, len(middle))
	for i := range middle {
		if i < period-1 {
			upper[i] = math.NaN()
			lower[i] = math.NaN()
		} else {
			upper[i] = middle[i] + stdDevMultiplier*stdDev[i]
			lower[i] = middle[i] - stdDevMultiplier*stdDev[i]
		}
	}

	return upper, middle, lower

}
