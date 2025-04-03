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

func CalculateEMA(candles []models.OHLCV, period int) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	ema := indicator.Ema(period, closes)
	return ema
}

func CalculateRSI(candles []models.OHLCV, period int) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	_, rsi := indicator.RsiPeriod(period, closes) // Используем RsiPeriod, игнорируем rs
	return rsi
}

func CalculateATR(candles []models.OHLCV, period int) []float64 {
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}
	_, atr := indicator.Atr(period, highs, lows, closes) // Игнорируем tr
	return atr
}

// CalculateADX временно закомментирован, так как Adx отсутствует в библиотеке
/*
func CalculateADX(candles []models.OHLCV, period int) []float64 {
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}
	// Здесь должна быть реализация ADX
	return make([]float64, len(candles)) // Заглушка
}
*/

func CalculateMACD(candles []models.OHLCV, shortPeriod, longPeriod, signalPeriod int) ([]float64, []float64) {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	macd, signal := indicator.Macd(closes) // Возвращаем только macd и signal
	return macd, signal
}
