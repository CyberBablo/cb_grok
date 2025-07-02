package indicators

import (
	candle_model "cb_grok/internal/models/candle"
	"github.com/cinar/indicator"
)

func CalculateEMA(candles []candle_model.OHLCV, period int) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	ema := indicator.Ema(period, closes)
	return ema
}
