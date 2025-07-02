package indicators

import (
	"cb_grok/pkg/models"
	"github.com/cinar/indicator"
)

func CalculateSMA(candles []models.OHLCV, period int) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	sma := indicator.Sma(period, closes)
	return sma
}

// calculateSMAFromK calculates SMA for %K values
func calculateSMAFromK(kValues []float64, period int) []float64 {
	sma := make([]float64, len(kValues))
	for i := period - 1; i < len(kValues); i++ {
		sum := 0.0
		count := 0
		for j := i - period + 1; j <= i; j++ {
			sum += kValues[j]
			count++
		}
		if count > 0 {
			sma[i] = sum / float64(count)
		}
	}
	return sma
}
