// Создаем новый файл: internal/ml/service.go
package ml

import (
	"cb_grok/pkg/models"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

type MLService struct {
	log          *zap.Logger
	Model        *Model
	Processor    *DataProcessor
	ModelPath    string
	FeaturesPath string
}

func NewMLService(log *zap.Logger) *MLService {
	// Создаем директорию для моделей, если ее нет
	modelsDir := filepath.Join("lib", "ml_models")
	if err := os.MkdirAll(modelsDir, os.ModePerm); err != nil {
		log.Error("Failed to create models directory", zap.Error(err))
	}

	// Параметры для ML
	windowSize := 10
	predictAhead := 3
	priceThreshold := 0.005 // 0.5% изменения цены для определения тренда
	treeCount := 100
	featureCount := 20 // Примерное количество признаков, которое мы извлекаем

	processor := NewDataProcessor(windowSize, predictAhead, priceThreshold)

	return &MLService{
		log:          log,
		Processor:    processor,
		ModelPath:    filepath.Join(modelsDir, "price_direction_model.gob"),
		FeaturesPath: filepath.Join(modelsDir, "features.json"),
		Model:        NewModel(treeCount, featureCount, processor),
	}
}

// TrainModel обучает модель на исторических данных
func (s *MLService) TrainModel(candles []models.AppliedOHLCV) error {
	s.log.Info("Training ML model...")

	// Обучаем модель
	if err := s.Model.Train(candles); err != nil {
		return err
	}

	// Сохраняем признаки для анализа (опционально)
	features := s.Processor.PrepareFeatures(candles)
	if err := s.Processor.SaveFeatures(features, s.FeaturesPath); err != nil {
		s.log.Warn("Failed to save features", zap.Error(err))
	}

	// Сохраняем модель
	if err := s.Model.Save(s.ModelPath); err != nil {
		return err
	}

	s.log.Info("ML model trained and saved successfully")
	return nil
}

// LoadOrTrainModel загружает существующую модель или обучает новую
func (s *MLService) LoadOrTrainModel(candles []models.AppliedOHLCV) error {
	// Пробуем загрузить модель
	model, err := LoadModel(s.ModelPath)
	if err == nil {
		s.log.Info("Loaded existing ML model")
		s.Model = model
		return nil
	}

	// Если не удалось загрузить, обучаем новую
	s.log.Info("No existing model found, training new model", zap.Error(err))
	return s.TrainModel(candles)
}

// Исправляем функцию PredictPriceDirection в файле internal/ml/service.go
func (s *MLService) PredictPriceDirection(candles []models.AppliedOHLCV) PriceDirection {
	if s.Model == nil {
		s.log.Error("Model not initialized")
		return Flat
	}

	// Проверка на наличие достаточного количества данных
	if len(candles) <= s.Processor.WindowSize {
		s.log.Warn("Not enough data for prediction",
			zap.Int("required", s.Processor.WindowSize+1),
			zap.Int("available", len(candles)))
		return Flat
	}

	return s.Model.Predict(candles)
}
