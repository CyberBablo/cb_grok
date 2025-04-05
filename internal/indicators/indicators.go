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
	if len(candles) < period+1 {
		return make([]float64, len(candles))
	}
	
	highs := make([]float64, len(candles))
	lows := make([]float64, len(candles))
	closes := make([]float64, len(candles))
	for i, c := range candles {
		highs[i] = c.High
		lows[i] = c.Low
		closes[i] = c.Close
	}
	
	tr := make([]float64, len(candles))
	tr[0] = highs[0] - lows[0] // First TR is just the range
	
	for i := 1; i < len(candles); i++ {
		tr[i] = max3(
			highs[i]-lows[i],
			math.Abs(highs[i]-closes[i-1]),
			math.Abs(lows[i]-closes[i-1]),
		)
	}
	
	plusDM := make([]float64, len(candles))
	minusDM := make([]float64, len(candles))
	
	for i := 1; i < len(candles); i++ {
		upMove := highs[i] - highs[i-1]
		downMove := lows[i-1] - lows[i]
		
		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		} else {
			plusDM[i] = 0
		}
		
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		} else {
			minusDM[i] = 0
		}
	}
	
	smoothTR := make([]float64, len(candles))
	smoothPlusDM := make([]float64, len(candles))
	smoothMinusDM := make([]float64, len(candles))
	
	for i := 1; i <= period; i++ {
		smoothTR[period] += tr[i]
		smoothPlusDM[period] += plusDM[i]
		smoothMinusDM[period] += minusDM[i]
	}
	
	for i := period + 1; i < len(candles); i++ {
		smoothTR[i] = smoothTR[i-1] - (smoothTR[i-1]/float64(period)) + tr[i]
		smoothPlusDM[i] = smoothPlusDM[i-1] - (smoothPlusDM[i-1]/float64(period)) + plusDM[i]
		smoothMinusDM[i] = smoothMinusDM[i-1] - (smoothMinusDM[i-1]/float64(period)) + minusDM[i]
	}
	
	plusDI := make([]float64, len(candles))
	minusDI := make([]float64, len(candles))
	
	for i := period; i < len(candles); i++ {
		if smoothTR[i] > 0 {
			plusDI[i] = 100 * (smoothPlusDM[i] / smoothTR[i])
			minusDI[i] = 100 * (smoothMinusDM[i] / smoothTR[i])
		}
	}
	
	dx := make([]float64, len(candles))
	
	for i := period; i < len(candles); i++ {
		if plusDI[i]+minusDI[i] > 0 {
			dx[i] = 100 * (math.Abs(plusDI[i]-minusDI[i]) / (plusDI[i] + minusDI[i]))
		}
	}
	
	adx := make([]float64, len(candles))
	
	sum := 0.0
	for i := period; i < 2*period; i++ {
		if i < len(dx) {
			sum += dx[i]
		}
	}
	if period < len(adx) {
		adx[2*period-1] = sum / float64(period)
	}
	
	for i := 2*period; i < len(candles); i++ {
		adx[i] = ((adx[i-1] * float64(period-1)) + dx[i]) / float64(period)
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

func max3(a, b, c float64) float64 {
	if a > b {
		if a > c {
			return a
		}
		return c
	}
	if b > c {
		return b
	}
	return c
}
