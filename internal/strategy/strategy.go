package strategy

import (
	candle_model "cb_grok/internal/models/candle"
	strategy_model "cb_grok/internal/models/strategy"
)

type Strategy interface {
	ApplyIndicators(candles []candle_model.OHLCV, params strategy_model.StrategyParamsModel) []candle_model.AppliedOHLCV
	ApplySignals(appliedCandles []candle_model.AppliedOHLCV, params strategy_model.StrategyParamsModel) []candle_model.AppliedOHLCV
}
