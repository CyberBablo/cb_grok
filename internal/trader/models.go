package trader

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/model"
	"cb_grok/internal/strategy"
	"cb_grok/pkg/models"
)

type TraderParams struct {
	Model          *model.Model
	Exchange       exchange.Exchange
	Strategy       strategy.Strategy
	Settings       *TraderSettings
	InitialCapital float64
}

type TraderSettings struct {
	Commission           float64
	SlippagePercent      float64
	Spread               float64
	StopLossMultiplier   float64
	TakeProfitMultiplier float64
}

type state struct {
	initialCapital float64

	ohlcv           []models.OHLCV
	appliedOHLCV    []models.AppliedOHLCV
	orders          []Action
	portfolioValues []PortfolioValue
}

type PortfolioValue struct {
	Timestamp int64
	Value     float64
}

type Action struct {
	Timestamp       int64
	Decision        TradeDecision
	DecisionTrigger TradeDecisionTrigger
	Price           float64
	AssetAmount     float64
	AssetCurrency   string
	Comment         string
	Profit          float64

	PortfolioValue float64
}
