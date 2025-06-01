package model

import (
	"cb_grok/internal/strategy"
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

type Model struct {
	Symbol string `json:"symbol"`
	strategy.StrategyParams
}

func Save(m Model) string {
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

func Load(filename string) (*Model, error) {
	path := filepath.Join(dir, filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m Model

	return &m, json.Unmarshal(data, &m)
}
