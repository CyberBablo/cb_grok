package strategy

import (
	"cb_grok/pkg/models"
	"math"
)

type Strategy interface {
	Apply(candles []models.OHLCV, params StrategyParams) []models.AppliedOHLCV
}

type StrategyParams struct {
	MAShortPeriod       int     `json:"ma_short_period,omitempty"`
	MALongPeriod        int     `json:"ma_long_period,omitempty"`
	RSIPeriod           int     `json:"rsi_period,omitempty"`
	ATRPeriod           int     `json:"atr_period,omitempty"`
	BuyRSIThreshold     float64 `json:"buy_rsi_threshold,omitempty"`
	SellRSIThreshold    float64 `json:"sell_rsi_threshold,omitempty"`
	EMAShortPeriod      int     `json:"ema_short_period,omitempty"`
	EMALongPeriod       int     `json:"ema_long_period,omitempty"`
	ATRThreshold        float64 `json:"atr_threshold,omitempty"`
	MACDShortPeriod     int     `json:"macd_short_period,omitempty"`
	MACDLongPeriod      int     `json:"macd_long_period,omitempty"`
	MACDSignalPeriod    int     `json:"macd_signal_period,omitempty"`
	EMAWeight           float64 `json:"ema_weight,omitempty"`
	TrendWeight         float64 `json:"trend_weight,omitempty"`
	RSIWeight           float64 `json:"rsi_weight,omitempty"`
	MACDWeight          float64 `json:"macd_weight,omitempty"`
	BuySignalThreshold  float64 `json:"buy_signal_threshold,omitempty"`
	SellSignalThreshold float64 `json:"sell_signal_threshold,omitempty"`
	BollingerPeriod     int     `json:"bollinger_period,omitempty"`
	BollingerStdDev     float64 `json:"bollinger_std_dev,omitempty"`
	BBWeight            float64 `json:"bb_weight,omitempty"`

	//Pair string `json:"pair,omitempty"`
}

type Signals struct {
	EMASignal   int
	RSISignal   int
	MACDSignal  int
	TrendSignal int
	BBSignal    int
}

func Tanh(x float64) float64 {
	return math.Tanh(x)
}
