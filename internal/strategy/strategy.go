package strategy

import (
	"cb_grok/pkg/models"
	strategy_model "cb_grok/internal/models/strategy"
)

type Strategy interface {
	ApplyIndicators(candles []models.OHLCV, params strategy_model.StrategyParamsModel) []models.AppliedOHLCV
	ApplySignals(appliedCandles []models.AppliedOHLCV, params strategy_model.StrategyParamsModel) []models.AppliedOHLCV
}
