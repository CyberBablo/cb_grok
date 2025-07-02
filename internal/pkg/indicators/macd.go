package indicators

import (
	candle_model "cb_grok/internal/models/candle"
	"github.com/cinar/indicator"
)

func CalculateMACD(candles []candle_model.OHLCV, shortPeriod, longPeriod, signalPeriod int) ([]float64, []float64) {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	macd, signal := indicator.Macd(closes) // Возвращаем только macd и signal
	return macd, signal
}
