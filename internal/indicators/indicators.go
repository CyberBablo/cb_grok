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
		smoothTR[i] = smoothTR[i-1] - (smoothTR[i-1] / float64(period)) + tr[i]
		smoothPlusDM[i] = smoothPlusDM[i-1] - (smoothPlusDM[i-1] / float64(period)) + plusDM[i]
		smoothMinusDM[i] = smoothMinusDM[i-1] - (smoothMinusDM[i-1] / float64(period)) + minusDM[i]
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
	for i := period; i < min(len(dx), 2*period); i++ {
		sum += dx[i]
	}
	
	if 2*period-1 < len(adx) && period > 0 {
		adx[2*period-1] = sum / float64(min(period, len(dx)-period))
	}

	for i := 2 * period; i < len(candles); i++ {
		if i-1 < len(adx) && i < len(dx) {
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func CalculateBollingerBands(candles []models.OHLCV, period int, stdDev float64) ([]float64, []float64, []float64) {
	if len(candles) < period {
		return make([]float64, len(candles)), make([]float64, len(candles)), make([]float64, len(candles))
	}

	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}

	sma := CalculateSMA(candles, period)
	
	upper := make([]float64, len(candles))
	lower := make([]float64, len(candles))
	
	for i := period - 1; i < len(candles); i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += math.Pow(closes[i-j]-sma[i], 2)
		}
		sd := math.Sqrt(sum / float64(period))
		
		upper[i] = sma[i] + stdDev*sd
		lower[i] = sma[i] - stdDev*sd
	}
	
	return upper, sma, lower
}

func CalculateStochastic(candles []models.OHLCV, kPeriod, dPeriod int) ([]float64, []float64) {
	if len(candles) < kPeriod {
		return make([]float64, len(candles)), make([]float64, len(candles))
	}
	
	k := make([]float64, len(candles))
	
	for i := kPeriod - 1; i < len(candles); i++ {
		highestHigh := candles[i].High
		lowestLow := candles[i].Low
		
		for j := 0; j < kPeriod; j++ {
			if candles[i-j].High > highestHigh {
				highestHigh = candles[i-j].High
			}
			if candles[i-j].Low < lowestLow {
				lowestLow = candles[i-j].Low
			}
		}
		
		if highestHigh - lowestLow > 0 {
			k[i] = 100 * ((candles[i].Close - lowestLow) / (highestHigh - lowestLow))
		} else {
			k[i] = 50 // Default to middle if range is zero
		}
	}
	
	d := make([]float64, len(candles))
	for i := kPeriod + dPeriod - 2; i < len(candles); i++ {
		sum := 0.0
		for j := 0; j < dPeriod; j++ {
			sum += k[i-j]
		}
		d[i] = sum / float64(dPeriod)
	}
	
	return k, d
}
