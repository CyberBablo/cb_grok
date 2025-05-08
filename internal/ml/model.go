// Создаем новый файл: internal/ml/model.go
package ml

import (
	"bytes"
	"cb_grok/pkg/models"
	"encoding/gob"
	"fmt"
	"math/rand"
	"os"
	"path"
	"time"
)

// PriceDirection представляет направление цены
type PriceDirection int

const (
	Down PriceDirection = -1
	Flat PriceDirection = 0
	Up   PriceDirection = 1
)

// Model представляет ML модель для предсказания
type Model struct {
	// В данной реализации мы будем использовать ансамбль деревьев решений
	Trees        []*DecisionTree
	FeatureCount int
	TreeCount    int

	// Данные для предобработки
	Processor *DataProcessor
}

// NewModel создает новую модель
func NewModel(treeCount, featureCount int, processor *DataProcessor) *Model {
	return &Model{
		Trees:        make([]*DecisionTree, treeCount),
		FeatureCount: featureCount,
		TreeCount:    treeCount,
		Processor:    processor,
	}
}

// Train обучает модель на исторических данных
func (m *Model) Train(candles []models.AppliedOHLCV) error {
	// Подготовка данных
	featureVectors := m.Processor.PrepareFeatures(candles)
	if len(featureVectors) == 0 {
		return fmt.Errorf("not enough data for training")
	}

	// Разделение на обучающую и валидационную выборку
	trainSize := int(float64(len(featureVectors)) * 0.8)
	trainData := featureVectors[:trainSize]
	validData := featureVectors[trainSize:]

	// Обучение каждого дерева
	for i := 0; i < m.TreeCount; i++ {
		// Создание случайной подвыборки для обучения (бутстрэп)
		bootstrapSample := m.bootstrap(trainData)

		// Создание и обучение дерева
		tree := NewDecisionTree(m.FeatureCount)
		tree.Train(bootstrapSample)

		m.Trees[i] = tree
	}

	// Оценка модели на валидационной выборке
	if len(validData) > 0 {
		accuracy := m.evaluate(validData)
		fmt.Printf("Model validation accuracy: %.2f%%\n", accuracy*100)
	}

	return nil
}

func (m *Model) Predict(candles []models.AppliedOHLCV) PriceDirection {
	// Проверка на достаточное количество данных
	if len(candles) <= m.Processor.WindowSize {
		return Flat // Недостаточно данных
	}

	// Извлекаем признаки для текущей точки (последняя свеча)
	features := m.Processor.extractFeatures(candles, len(candles)-1)

	// Проверяем, что получили непустой вектор признаков
	if len(features) == 0 {
		return Flat // Не удалось извлечь признаки
	}

	// Голосование деревьев
	votes := make([]float64, 3) // Down, Flat, Up

	for _, tree := range m.Trees {
		prediction := tree.Predict(features)

		// Преобразуем предсказанное значение в индекс
		index := int(prediction + 1) // -1 -> 0, 0 -> 1, 1 -> 2
		if index >= 0 && index < 3 {
			votes[index]++
		}
	}

	// Определяем победителя голосования
	maxVotes := votes[0]
	maxIndex := 0

	for i := 1; i < len(votes); i++ {
		if votes[i] > maxVotes {
			maxVotes = votes[i]
			maxIndex = i
		}
	}

	return PriceDirection(maxIndex - 1) // 0 -> -1, 1 -> 0, 2 -> 1
}

// bootstrap создает случайную подвыборку с возвращением
func (m *Model) bootstrap(data []FeatureVector) []FeatureVector {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := make([]FeatureVector, len(data))

	for i := 0; i < len(data); i++ {
		idx := rng.Intn(len(data))
		result[i] = data[idx]
	}

	return result
}

// evaluate оценивает точность модели
func (m *Model) evaluate(validData []FeatureVector) float64 {
	if len(validData) == 0 {
		return 0.0
	}

	correct := 0

	for _, sample := range validData {
		// Голосование деревьев
		votes := make([]float64, 3) // Down, Flat, Up

		for _, tree := range m.Trees {
			prediction := tree.Predict(sample.Features)

			// Преобразуем предсказанное значение в индекс
			index := int(prediction + 1) // -1 -> 0, 0 -> 1, 1 -> 2
			if index >= 0 && index < 3 {
				votes[index]++
			}
		}

		// Определяем победителя голосования
		maxVotes := votes[0]
		maxIndex := 0

		for i := 1; i < len(votes); i++ {
			if votes[i] > maxVotes {
				maxVotes = votes[i]
				maxIndex = i
			}
		}

		prediction := PriceDirection(maxIndex - 1) // 0 -> -1, 1 -> 0, 2 -> 1

		if float64(prediction) == sample.Label {
			correct++
		}
	}

	return float64(correct) / float64(len(validData))
}

// Save сохраняет модель в файл
func (m *Model) Save(filepath string) error {
	// Создаем директорию, если она не существует
	dir := path.Dir(filepath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	buffer := bytes.Buffer{}
	encoder := gob.NewEncoder(&buffer)

	if err := encoder.Encode(m); err != nil {
		return fmt.Errorf("failed to encode model: %w", err)
	}

	if _, err := file.Write(buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// Load загружает модель из файла
func LoadModel(filepath string) (*Model, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(file); err != nil {
		return nil, fmt.Errorf("failed to read from file: %w", err)
	}

	decoder := gob.NewDecoder(&buffer)
	var model Model

	if err := decoder.Decode(&model); err != nil {
		return nil, fmt.Errorf("failed to decode model: %w", err)
	}

	return &model, nil
}
