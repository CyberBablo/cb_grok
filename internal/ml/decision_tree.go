package ml

import (
	"math"
	"math/rand"
	"time"
)

// DecisionTree представляет дерево решений
type DecisionTree struct {
	Root           *TreeNode
	MaxDepth       int
	MinSampleSplit int
	FeatureCount   int
}

// TreeNode представляет узел дерева решений
type TreeNode struct {
	FeatureIndex int       // Индекс признака для разделения
	Threshold    float64   // Порог для разделения
	Left         *TreeNode // Левый потомок
	Right        *TreeNode // Правый потомок
	Value        float64   // Значение (для листового узла)
	IsLeaf       bool      // Является ли узел листовым
}

// NewDecisionTree создает новое дерево решений
func NewDecisionTree(featureCount int) *DecisionTree {
	return &DecisionTree{
		MaxDepth:       5,            // Максимальная глубина дерева
		MinSampleSplit: 5,            // Минимальное количество образцов для разделения
		FeatureCount:   featureCount, // Количество признаков
	}
}

// Train обучает дерево решений
func (dt *DecisionTree) Train(data []FeatureVector) {
	if len(data) == 0 {
		return
	}

	// Создаем корневой узел
	dt.Root = dt.buildTree(data, 0)
}

// buildTree рекурсивно строит дерево
func (dt *DecisionTree) buildTree(data []FeatureVector, depth int) *TreeNode {
	// Проверяем условия остановки
	if depth >= dt.MaxDepth || len(data) < dt.MinSampleSplit {
		return &TreeNode{
			IsLeaf: true,
			Value:  dt.calculateLeafValue(data),
		}
	}

	// Находим лучшее разделение
	bestFeatureIndex, bestThreshold := dt.findBestSplit(data)

	// Если не удалось найти хорошее разделение, создаем лист
	if bestFeatureIndex == -1 {
		return &TreeNode{
			IsLeaf: true,
			Value:  dt.calculateLeafValue(data),
		}
	}

	// Разделяем данные
	leftData, rightData := dt.splitData(data, bestFeatureIndex, bestThreshold)

	// Если одно из разделений пустое, создаем лист
	if len(leftData) == 0 || len(rightData) == 0 {
		return &TreeNode{
			IsLeaf: true,
			Value:  dt.calculateLeafValue(data),
		}
	}

	// Создаем узел
	node := &TreeNode{
		FeatureIndex: bestFeatureIndex,
		Threshold:    bestThreshold,
		IsLeaf:       false,
	}

	// Рекурсивно строим левое и правое поддерево
	node.Left = dt.buildTree(leftData, depth+1)
	node.Right = dt.buildTree(rightData, depth+1)

	return node
}

// findBestSplit находит лучшее разделение
func (dt *DecisionTree) findBestSplit(data []FeatureVector) (int, float64) {
	if len(data) == 0 {
		return -1, 0.0
	}

	bestGain := 0.0
	bestFeatureIndex := -1
	bestThreshold := 0.0

	// Выбираем случайное подмножество признаков (Random Forest)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	featureIndices := make([]int, 0)

	// Выбираем sqrt(n) признаков из общего числа
	numFeatures := int(math.Sqrt(float64(dt.FeatureCount)))
	if numFeatures < 1 {
		numFeatures = 1
	}

	for len(featureIndices) < numFeatures && len(featureIndices) < dt.FeatureCount {
		idx := rng.Intn(dt.FeatureCount)
		// Проверяем, что признак еще не выбран
		alreadySelected := false
		for _, i := range featureIndices {
			if i == idx {
				alreadySelected = true
				break
			}
		}
		if !alreadySelected {
			featureIndices = append(featureIndices, idx)
		}
	}

	// Для каждого признака находим лучший порог
	for _, featureIndex := range featureIndices {
		// Собираем значения признака
		featureValues := make([]float64, len(data))
		for i, sample := range data {
			if featureIndex < len(sample.Features) {
				featureValues[i] = sample.Features[featureIndex]
			}
		}

		// Находим уникальные значения признака (или подмножество значений)
		uniqueValues := make([]float64, 0)
		for _, value := range featureValues {
			isUnique := true
			for _, unique := range uniqueValues {
				if math.Abs(value-unique) < 1e-6 {
					isUnique = false
					break
				}
			}
			if isUnique {
				uniqueValues = append(uniqueValues, value)
			}
		}

		// Ограничиваем количество значений для ускорения
		if len(uniqueValues) > 10 {
			subset := make([]float64, 10)
			for i := 0; i < 10; i++ {
				subset[i] = uniqueValues[rng.Intn(len(uniqueValues))]
			}
			uniqueValues = subset
		}

		// Для каждого уникального значения пробуем создать порог
		for _, threshold := range uniqueValues {
			// Разделяем данные
			leftData, rightData := dt.splitData(data, featureIndex, threshold)

			// Если одно из разделений пустое, пропускаем
			if len(leftData) == 0 || len(rightData) == 0 {
				continue
			}

			// Вычисляем информационный выигрыш
			gain := dt.calculateGain(data, leftData, rightData)

			// Если выигрыш лучше, запоминаем
			if gain > bestGain {
				bestGain = gain
				bestFeatureIndex = featureIndex
				bestThreshold = threshold
			}
		}
	}

	return bestFeatureIndex, bestThreshold
}

// splitData разделяет данные по порогу
func (dt *DecisionTree) splitData(data []FeatureVector, featureIndex int, threshold float64) ([]FeatureVector, []FeatureVector) {
	leftData := make([]FeatureVector, 0)
	rightData := make([]FeatureVector, 0)

	for _, sample := range data {
		if featureIndex < len(sample.Features) && sample.Features[featureIndex] <= threshold {
			leftData = append(leftData, sample)
		} else {
			rightData = append(rightData, sample)
		}
	}

	return leftData, rightData
}

// calculateGain вычисляет информационный выигрыш
func (dt *DecisionTree) calculateGain(parent, left, right []FeatureVector) float64 {
	if len(parent) == 0 || len(left) == 0 || len(right) == 0 {
		return 0.0
	}

	parentImpurity := dt.calculateImpurity(parent)
	leftWeight := float64(len(left)) / float64(len(parent))
	rightWeight := float64(len(right)) / float64(len(parent))

	gain := parentImpurity
	gain -= leftWeight * dt.calculateImpurity(left)
	gain -= rightWeight * dt.calculateImpurity(right)

	return gain
}

// calculateImpurity вычисляет неоднородность
func (dt *DecisionTree) calculateImpurity(data []FeatureVector) float64 {
	if len(data) == 0 {
		return 0.0
	}

	// Используем MSE для регрессии
	sum := 0.0
	for _, sample := range data {
		sum += sample.Label
	}
	mean := sum / float64(len(data))

	mse := 0.0
	for _, sample := range data {
		mse += (sample.Label - mean) * (sample.Label - mean)
	}

	return mse / float64(len(data))
}

// calculateLeafValue вычисляет значение листового узла
func (dt *DecisionTree) calculateLeafValue(data []FeatureVector) float64 {
	if len(data) == 0 {
		return 0.0
	}

	// Среднее значение целевой переменной
	sum := 0.0
	for _, sample := range data {
		sum += sample.Label
	}

	return sum / float64(len(data))
}

// Predict предсказывает значение
func (dt *DecisionTree) Predict(features []float64) float64 {
	if dt.Root == nil {
		return 0.0
	}

	return dt.predictNode(dt.Root, features)
}

// predictNode предсказывает значение для узла
func (dt *DecisionTree) predictNode(node *TreeNode, features []float64) float64 {
	if node.IsLeaf {
		return node.Value
	}

	if node.FeatureIndex < len(features) && features[node.FeatureIndex] <= node.Threshold {
		return dt.predictNode(node.Left, features)
	} else {
		return dt.predictNode(node.Right, features)
	}
}
