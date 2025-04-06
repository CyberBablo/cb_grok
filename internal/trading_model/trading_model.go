package trading_model

import (
	"cb_grok/internal/strategy"
	"cb_grok/internal/utils/logger"
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

var (
	log = logger.GetInstance()
)

func SaveModel(symbol string, params strategy.StrategyParams) string {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Error("create dir", zap.Error(err))
		return ""
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := filepath.Join(dir, fmt.Sprintf("model_%s.json", timestamp))
	file, err := os.Create(filename)
	if err != nil {
		log.Error("create new model file", zap.Error(err))
		return ""
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	var m map[string]interface{}

	b, _ := json.Marshal(params)
	_ = json.Unmarshal(b, &m)

	m["symbol"] = symbol

	if err := encoder.Encode(m); err != nil {
		log.Error("optimize: model file encoder", zap.Error(err))
		return ""
	}

	return filename
}
