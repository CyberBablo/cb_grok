package indicators

import "cb_grok/pkg/models"

// CalculateVWAP calculates the Volume Weighted Average Price for the given OHLCV data over a specified period.
func CalculateVWAP(candles []models.OHLCV, period int) []float64 {
	if len(candles) < period {
		return nil
	}

	vwap := make([]float64, len(candles))
	var sumVolume float64
	var sumPriceVolume float64

	for i := 0; i < period; i++ {
		typicalPrice := (candles[i].High + candles[i].Low + candles[i].Close) / 3
		sumPriceVolume += typicalPrice * float64(candles[i].Volume)
		sumVolume += float64(candles[i].Volume)
		vwap[i] = sumPriceVolume / sumVolume
	}

	for i := period; i < len(candles); i++ {
		typicalPrice := (candles[i].High + candles[i].Low + candles[i].Close) / 3
		sumPriceVolume += typicalPrice * float64(candles[i].Volume)
		sumPriceVolume -= (candles[i-period].High + candles[i-period].Low + candles[i-period].Close) / 3 * float64(candles[i-period].Volume)
		sumVolume += float64(candles[i].Volume)
		sumVolume -= float64(candles[i-period].Volume)
		vwap[i] = sumPriceVolume / sumVolume
	}

	return vwap
}
