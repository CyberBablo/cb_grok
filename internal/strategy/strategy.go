package strategy

import (
	strategy_model "cb_grok/internal/models/strategy"
	"cb_grok/pkg/models"
)

type Strategy interface {
	ApplyIndicators(candles []models.OHLCV, params strategy_model.StrategyParamsModel) []models.AppliedOHLCV
	ApplySignals(appliedCandles []models.AppliedOHLCV, params strategy_model.StrategyParamsModel) []models.AppliedOHLCV
}
