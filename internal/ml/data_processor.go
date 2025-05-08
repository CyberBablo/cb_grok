package ml

import (
	"cb_grok/pkg/models"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/samber/lo"
)

type FeatureVector struct {
	Features []float64
	Label    float64 // 1.0 для роста цены, -1.0 для падения, 0.0 для стабильной
}

type DataProcessor struct {
	WindowSize     int     // Размер окна для анализа тренда
	PredictAhead   int     // На сколько свечей вперед предсказываем
	PriceThreshold float64 // Порог изменения цены для определения тренда
	FeatureCount   int
}

func NewDataProcessor(windowSize, predictAhead int, priceThreshold float64) *DataProcessor {
	return &DataProcessor{
		WindowSize:     windowSize,
		PredictAhead:   predictAhead,
		PriceThreshold: priceThreshold,
		FeatureCount:   20,
	}
}

// PrepareFeatures готовит признаки из исторических данных
func (dp *DataProcessor) PrepareFeatures(candles []models.AppliedOHLCV) []FeatureVector {
	if len(candles) < dp.WindowSize+dp.PredictAhead {
		return nil
	}

	var featureVectors []FeatureVector

	// Создаем векторы признаков для каждой точки данных, которая может быть использована
	for i := dp.WindowSize; i < len(candles)-dp.PredictAhead; i++ {
		// Создаем вектор признаков
		features := dp.extractFeatures(candles, i)

		// Определяем метку (направление цены)
		label := dp.extractLabel(candles, i)

		featureVectors = append(featureVectors, FeatureVector{
			Features: features,
			Label:    label,
		})
	}

	return featureVectors
}

// extractFeatures извлекает признаки для машинного обучения
func (dp *DataProcessor) extractFeatures(candles []models.AppliedOHLCV, currentIndex int) []float64 {
	var features []float64

	// Проверка на корректность индекса
	if currentIndex < dp.WindowSize || currentIndex >= len(candles) {
		// Возвращаем пустой вектор признаков в случае некорректного индекса
		// Это предотвратит панику, но потом нужно обработать этот случай в коде вызова
		return make([]float64, dp.FeatureCount) // Возвращаем нулевой вектор нужной длины
	}

	// Используем окно для создания признаков
	window := candles[currentIndex-dp.WindowSize : currentIndex]
	current := candles[currentIndex]

	// 1. Технические индикаторы (текущие значения)
	features = append(features, current.RSI/100.0) // Нормализуем RSI
	features = append(features, current.MACD)
	features = append(features, current.ShortMA/current.Close) // Относительные значения
	features = append(features, current.LongMA/current.Close)
	features = append(features, current.ShortEMA/current.Close)
	features = append(features, current.LongEMA/current.Close)
	features = append(features, lo.If(current.Trend, 1.0).Else(0.0))
	features = append(features, lo.If(current.Volatility, 1.0).Else(0.0))
	features = append(features, current.ATR/current.Close) // Нормализуем ATR
	features = append(features, current.StochasticK/100.0) // Нормализуем Stochastic
	features = append(features, current.StochasticD/100.0)

	// 3. Изменение объема
	if len(window) > 1 {
		avgVolume := 0.0
		for _, c := range window[:len(window)-1] {
			avgVolume += c.Volume
		}
		avgVolume /= float64(len(window) - 1)
		if avgVolume > 0 {
			features = append(features, current.Volume/avgVolume-1.0)
		} else {
			features = append(features, 0.0)
		}
	} else {
		features = append(features, 0.0)
	}

	// 4. Тренды и паттерны из окна
	// Локальный тренд (линейная регрессия угла наклона)
	slope := dp.calculatePriceSlope(window)
	features = append(features, slope)

	// 5. Статистические характеристики свечей в окне
	closePrices := make([]float64, len(window))
	for i, c := range window {
		closePrices[i] = c.Close
	}
	mean, stdDev := dp.calculateStats(closePrices)
	features = append(features, mean/current.Close)
	features = append(features, stdDev/current.Close)

	// 6. Ценовые уровни от Bollinger Bands
	if !math.IsNaN(current.UpperBB) && !math.IsNaN(current.LowerBB) && current.UpperBB != current.LowerBB {
		bbPosition := (current.Close - current.LowerBB) / (current.UpperBB - current.LowerBB)
		features = append(features, bbPosition)
	} else {
		features = append(features, 0.5) // Середина по умолчанию
	}

	// 7. Временные признаки (часы, день недели и т.д.)
	t := time.UnixMilli(current.Timestamp)
	hour := float64(t.Hour()) / 24.0
	dayOfWeek := float64(t.Weekday()) / 7.0
	features = append(features, hour)
	features = append(features, dayOfWeek)

	return features
}

// calculatePriceSlope рассчитывает угол наклона линии тренда
func (dp *DataProcessor) calculatePriceSlope(window []models.AppliedOHLCV) float64 {
	if len(window) < 2 {
		return 0.0
	}

	n := float64(len(window))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, c := range window {
		x := float64(i)
		y := c.Close

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Формула линейной регрессии для угла наклона
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return 0.0
	}

	slope := (n*sumXY - sumX*sumY) / denominator

	// Нормализуем относительно последней цены
	return slope / window[len(window)-1].Close
}

// calculateStats рассчитывает среднее и стандартное отклонение
func (dp *DataProcessor) calculateStats(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0.0, 0.0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	sumSquaredDiff := 0.0
	for _, v := range values {
		sumSquaredDiff += math.Pow(v-mean, 2)
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(len(values)))

	return mean, stdDev
}

// extractLabel определяет метку (направление цены)
func (dp *DataProcessor) extractLabel(candles []models.AppliedOHLCV, currentIndex int) float64 {
	if currentIndex+dp.PredictAhead >= len(candles) {
		return 0.0
	}

	current := candles[currentIndex].Close
	future := candles[currentIndex+dp.PredictAhead].Close

	priceChange := (future - current) / current

	if priceChange > dp.PriceThreshold {
		return 1.0 // Рост
	} else if priceChange < -dp.PriceThreshold {
		return -1.0 // Падение
	}

	return 0.0 // Боковое движение
}

// SaveFeatures сохраняет признаки в файл для анализа
func (dp *DataProcessor) SaveFeatures(features []FeatureVector, filename string) error {
	// Убедимся, что директория существует
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(features)
}
