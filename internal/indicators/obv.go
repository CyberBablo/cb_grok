package indicators

import (
	"cb_grok/pkg/models"
	"github.com/cinar/indicator"
)

// CalculateOBV calculates the On-Balance Volume for the given OHLCV data.
func CalculateOBV(candles []models.OHLCV) []float64 {
	closes := make([]float64, len(candles))
	volumes := make([]float64, len(candles))

	for i, c := range candles {
		closes[i] = c.Close
		volumes[i] = c.Volume
	}

	obv := indicator.Obv(closes, volumes)
	return obv
}
