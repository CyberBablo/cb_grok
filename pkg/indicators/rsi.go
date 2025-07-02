package indicators

import (
	candle_model "cb_grok/internal/models/candle"
	"github.com/cinar/indicator"
)

func CalculateRSI(candles []candle_model.OHLCV, period int) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	_, rsi := indicator.RsiPeriod(period, closes) // Используем RsiPeriod, игнорируем rs
	return rsi
}
