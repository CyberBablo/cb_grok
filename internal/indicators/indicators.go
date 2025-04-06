package indicators

import (
	"cb_grok/pkg/models"
	"github.com/cinar/indicator"
	"math"
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

func CalculateADX(candles []models.OHLCV, period int) []float64 {
	n := len(candles)
	if n <= period {
		return make([]float64, n) // Return zeros if not enough data
	}

	tr := make([]float64, n)          // True Range
	posDM := make([]float64, n)       // Positive Directional Movement
	negDM := make([]float64, n)       // Negative Directional Movement
	smoothPosDM := make([]float64, n) // Smoothed +DM
	smoothNegDM := make([]float64, n) // Smoothed -DM
	smoothTR := make([]float64, n)    // Smoothed TR
	posDI := make([]float64, n)       // Positive Directional Indicator
	negDI := make([]float64, n)       // Negative Directional Indicator
	dx := make([]float64, n)          // Directional Index
	adx := make([]float64, n)         // Average Directional Index

	for i := 1; i < n; i++ {
		tr[i] = math.Max(candles[i].High-candles[i].Low, math.Max(
			math.Abs(candles[i].High-candles[i-1].Close),
			math.Abs(candles[i].Low-candles[i-1].Close)))

		upMove := candles[i].High - candles[i-1].High
		downMove := candles[i-1].Low - candles[i].Low

		if upMove > downMove && upMove > 0 {
			posDM[i] = upMove
		} else {
			posDM[i] = 0
		}

		if downMove > upMove && downMove > 0 {
			negDM[i] = downMove
		} else {
			negDM[i] = 0
		}
	}

	for i := 1; i <= period; i++ {
		if i == period {
			smoothPosDM[i] = 0
			smoothNegDM[i] = 0
			smoothTR[i] = 0

			for j := 1; j <= period; j++ {
				smoothPosDM[i] += posDM[j]
				smoothNegDM[i] += negDM[j]
				smoothTR[i] += tr[j]
			}
		}
	}

	for i := period + 1; i < n; i++ {
		smoothPosDM[i] = smoothPosDM[i-1] - (smoothPosDM[i-1] / float64(period)) + posDM[i]
		smoothNegDM[i] = smoothNegDM[i-1] - (smoothNegDM[i-1] / float64(period)) + negDM[i]
		smoothTR[i] = smoothTR[i-1] - (smoothTR[i-1] / float64(period)) + tr[i]
	}

	for i := period; i < n; i++ {
		if smoothTR[i] > 0 {
			posDI[i] = 100 * (smoothPosDM[i] / smoothTR[i])
			negDI[i] = 100 * (smoothNegDM[i] / smoothTR[i])
		}
	}

	for i := period; i < n; i++ {
		if posDI[i]+negDI[i] > 0 {
			dx[i] = 100 * (math.Abs(posDI[i]-negDI[i]) / (posDI[i] + negDI[i]))
		}
	}

	for i := 2*period - 1; i < n; i++ {
		if i == 2*period-1 {
			adx[i] = 0
			for j := period; j <= 2*period-1; j++ {
				adx[i] += dx[j]
			}
			adx[i] = adx[i] / float64(period)
		} else {
			adx[i] = ((adx[i-1] * float64(period-1)) + dx[i]) / float64(period)
		}
	}

	return adx
}

func CalculateMACD(candles []models.OHLCV, shortPeriod, longPeriod, signalPeriod int) ([]float64, []float64) {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	macd, signal := indicator.Macd(closes) // Возвращаем только macd и signal
	return macd, signal
}
