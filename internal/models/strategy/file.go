package strategy

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

const (
	dir = "lib/best_models"
)

type StrategyFileModel struct {
	Symbol string `json:"symbol"`
	StrategyParamsModel
}

func Save(m StrategyFileModel) string {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		zap.L().Error("create dir", zap.Error(err))
		return ""
	}

	timestamp := time.Now().Format("20060102_150405")
	path := filepath.Join(dir, fmt.Sprintf("model_%s.json", timestamp))
	file, err := os.Create(path)
	if err != nil {
		zap.L().Error("create new model file", zap.Error(err))
		return ""
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(m); err != nil {
		zap.L().Error("model file encoder", zap.Error(err))
		return ""
	}

	return path
}

func Load(filename string) (*StrategyFileModel, error) {
	path := filepath.Join(dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m StrategyFileModel

	return &m, json.Unmarshal(data, &m)
}
