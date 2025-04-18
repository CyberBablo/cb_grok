package indicators

import (
	"cb_grok/pkg/models"
	"github.com/cinar/indicator"
)

// CalculateCCI calculates the Commodity Channel Index for the given OHLCV data over a specified period.
func CalculateCCI(candles []models.OHLCV, period int) []float64 {
	closes := make([]float64, len(candles))
	volumes := make([]float64, len(candles))
	high := make([]float64, len(candles))
	low := make([]float64, len(candles))

	for i, c := range candles {
		closes[i] = c.Close
		high[i] = c.High
		low[i] = c.Low
		volumes[i] = c.Volume
	}

	cci := indicator.CommunityChannelIndex(period, high, low, closes)
	return cci
}
