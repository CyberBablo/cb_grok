package indicators

import (
	"cb_grok/pkg/models"
	"math"
)

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
