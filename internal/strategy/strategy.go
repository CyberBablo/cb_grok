package strategy

import (
	"cb_grok/pkg/models"
)

type Strategy interface {
	Apply(candles []models.OHLCV, params StrategyParams) []models.AppliedOHLCV
}

type StrategyParams struct {
	MAShortPeriod        int
	MALongPeriod         int
	RSIPeriod            int
	ATRPeriod            int
	BuyRSIThreshold      float64
	SellRSIThreshold     float64
	StopLossMultiplier   float64
	TakeProfitMultiplier float64
	EMAShortPeriod       int
	EMALongPeriod        int
	UseTrendFilter       bool
	UseRSIFilter         bool
	ATRThreshold         float64
	MACDShort            int
	MACDLong             int
	MACDSignal           int
	BollingerPeriod      int
	BollingerStdDev      float64
	UseMACDFilter        bool
	UseBollingerFilter   bool
}
