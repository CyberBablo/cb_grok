package strategy

import (
	"cb_grok/pkg/models"
)

type Strategy interface {
	Apply(candles []models.OHLCV, params StrategyParams) []models.AppliedOHLCV
}

type StrategyParams struct {
	MAShortPeriod        int     `json:"ma_short_period"`
	MALongPeriod         int     `json:"ma_long_period"`
	RSIPeriod            int     `json:"rsi_period"`
	ATRPeriod            int     `json:"atr_period"`
	BuyRSIThreshold      float64 `json:"buy_rsi_threshold"`
	SellRSIThreshold     float64 `json:"sell_rsi_threshold"`
	StopLossMultiplier   float64 `json:"stop_loss_multiplier"`
	TakeProfitMultiplier float64 `json:"take_profit_multiplier"`
	EMAShortPeriod       int     `json:"ema_short_period"`
	EMALongPeriod        int     `json:"ema_long_period"`
	UseTrendFilter       bool    `json:"use_trend_filter"`
	UseRSIFilter         bool    `json:"use_rsi_filter"`
	ATRThreshold         float64 `json:"atr_threshold"`
	ADXPeriod            int     `json:"adx_period"`
	ADXThreshold         float64 `json:"adx_threshold"`
	MACDShortPeriod      int     `json:"macd_short_period"`
	MACDLongPeriod       int     `json:"macd_long_period"`
	MACDSignalPeriod     int     `json:"macd_signal_period"`
}
