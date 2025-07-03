package strategy

import (
	"cb_grok/internal/strategy/model"
	"cb_grok/pkg/models"
)

type Strategy interface {
	ApplyIndicators(candles []models.OHLCV, params model.StrategyParams) []models.AppliedOHLCV
	ApplySignals(appliedCandles []models.AppliedOHLCV, params model.StrategyParams) []models.AppliedOHLCV
}
