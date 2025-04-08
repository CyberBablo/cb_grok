package indicators

import (
	"cb_grok/pkg/models"
	"math"
)

// CalculateStochasticOscillator calculates the Stochastic Oscillator (%K and %D)
func CalculateStochasticOscillator(candles []models.OHLCV, kPeriod int, dPeriod int) ([]float64, []float64) {
	if len(candles) < kPeriod {
		return nil, nil
	}

	kValues := make([]float64, len(candles))
	dValues := make([]float64, len(candles))

	// Расчет %K
	for i := kPeriod - 1; i < len(candles); i++ {
		high := math.Inf(-1)
		low := math.Inf(1)
		for j := i - kPeriod + 1; j <= i; j++ {
			if candles[j].High > high {
				high = candles[j].High
			}
			if candles[j].Low < low {
				low = candles[j].Low
			}
		}
		c := candles[i].Close
		if high == low { // Избегаем деления на ноль
			kValues[i] = 50.0
		} else {
			kValues[i] = (c - low) / (high - low) * 100
		}
	}

	// Расчет %D как SMA от %K
	dValues = calculateSMAFromK(kValues, dPeriod)

	return kValues, dValues
}
