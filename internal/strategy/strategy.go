package strategy

import "cb_grok/pkg/models"

type Strategy interface {
	ApplyIndicators(candles []models.OHLCV, params StrategyParams) []models.AppliedOHLCV
	ApplySignals(appliedCandles []models.AppliedOHLCV, params StrategyParams) []models.AppliedOHLCV
}
