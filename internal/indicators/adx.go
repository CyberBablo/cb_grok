package indicators

import (
	"cb_grok/pkg/models"
	"math"
)

// CalculateADX вычисляет ADX на основе массива свечей и заданного периода.
func CalculateADX(candles []models.Candle, period int) []float64 {
	if len(candles) < period+1 {
		return nil // Недостаточно данных для расчета
	}

	// Шаг 1: Рассчитываем True Range (TR), +DM и -DM
	tr := make([]float64, len(candles))
	plusDM := make([]float64, len(candles))
	minusDM := make([]float64, len(candles))

	for i := 1; i < len(candles); i++ {
		high := candles[i].High
		low := candles[i].Low
		prevClose := candles[i-1].Close

		// True Range
		hMinusL := high - low
		hMinusPrevC := math.Abs(high - prevClose)
		lMinusPrevC := math.Abs(low - prevClose)
		tr[i] = math.Max(hMinusL, math.Max(hMinusPrevC, lMinusPrevC))

		// +DM и -DM
		upMove := high - candles[i-1].High
		downMove := candles[i-1].Low - low
		if upMove > downMove && upMove > 0 {
			plusDM[i] = upMove
		}
		if downMove > upMove && downMove > 0 {
			minusDM[i] = downMove
		}
	}

	// Шаг 2: Сглаживаем TR, +DM и -DM для получения ATR, +DI и -DI
	atr := smooth(tr, period)
	smoothedPlusDM := smooth(plusDM, period)
	smoothedMinusDM := smooth(minusDM, period)

	// Шаг 3: Рассчитываем +DI и -DI
	plusDI := make([]float64, len(candles))
	minusDI := make([]float64, len(candles))
	for i := period; i < len(candles); i++ {
		if atr[i] != 0 {
			plusDI[i] = 100 * smoothedPlusDM[i] / atr[i]
			minusDI[i] = 100 * smoothedMinusDM[i] / atr[i]
		}
	}

	// Шаг 4: Рассчитываем DX (Directional Index)
	dx := make([]float64, len(candles))
	for i := period; i < len(candles); i++ {
		diSum := plusDI[i] + minusDI[i]
		diDiff := math.Abs(plusDI[i] - minusDI[i])
		if diSum != 0 {
			dx[i] = 100 * diDiff / diSum
		}
	}

	// Шаг 5: Сглаживаем DX для получения ADX
	adx := smooth(dx, period)
	return adx
}

// smooth выполняет сглаживание массива с использованием Wilder's Moving Average.
func smooth(data []float64, period int) []float64 {
	result := make([]float64, len(data))
	if len(data) < period {
		return result
	}

	// Инициализация: простая средняя для первых `period` значений
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[period-1] = sum / float64(period)

	// Сглаживание по методу Уайлдера
	for i := period; i < len(data); i++ {
		result[i] = (result[i-1]*float64(period-1) + data[i]) / float64(period)
	}
	return result
}
