package strategy

import (
	"cb_grok/pkg/models"
)

type Strategy interface {
	Apply(candles []models.OHLCV, params StrategyParams) []models.AppliedOHLCV
}
