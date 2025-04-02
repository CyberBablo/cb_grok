package strategy

import (
	"cb_grok/internal/indicators"
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
}

type strategyImpl struct{}

func NewStrategy() Strategy {
	return &strategyImpl{}
}

func (s *strategyImpl) Apply(candles []models.OHLCV, params StrategyParams) []models.AppliedOHLCV {
	if len(candles) < max(params.MALongPeriod, params.EMALongPeriod, 0) {
		return nil // Недостаточно данных
	}

	shortMA := indicators.CalculateSMA(candles, params.MAShortPeriod)
	longMA := indicators.CalculateSMA(candles, params.MALongPeriod)
	rsi := indicators.CalculateRSI(candles, params.RSIPeriod)
	atr := indicators.CalculateATR(candles, params.ATRPeriod)
	emaShort := indicators.CalculateEMA(candles, params.EMAShortPeriod)
	emaLong := indicators.CalculateEMA(candles, params.EMALongPeriod)

	//var adx []float64
	//if params.UseADXFilter {
	//	adx = indicators.CalculateADX(candles, params.ADXPeriod)
	//}

	trend := make([]bool, len(candles))
	volatility := make([]bool, len(candles))
	for i := range candles {
		trend[i] = emaShort[i] > emaLong[i]
		volatility[i] = atr[i] > params.ATRThreshold
	}

	var appliedCandles []models.AppliedOHLCV
	for i := range candles {
		appliedCandles = append(appliedCandles, models.AppliedOHLCV{
			OHLCV:      candles[i],
			ATR:        atr[i],
			RSI:        rsi[i],
			ShortMA:    shortMA[i],
			LongMA:     longMA[i],
			ShortEMA:   emaShort[i],
			LongEMA:    emaLong[i],
			Trend:      trend[i],
			Volatility: volatility[i],
		})
	}

	for i := 1; i < len(appliedCandles); i++ {
		buyCondition := shortMA[i] > longMA[i]
		sellCondition := shortMA[i] < longMA[i]

		if params.UseRSIFilter {
			buyCondition = buyCondition && rsi[i] < params.BuyRSIThreshold
			sellCondition = sellCondition && rsi[i] > params.SellRSIThreshold
		}

		//if params.UseADXFilter && adx != nil {
		//	buyCondition = buyCondition && adx[i] > params.ADXThreshold
		//	sellCondition = sellCondition && adx[i] > params.ADXThreshold
		//}

		if params.UseTrendFilter {
			buyCondition = buyCondition && trend[i] && volatility[i]
			sellCondition = sellCondition && !trend[i] && volatility[i]
		}

		if buyCondition {
			appliedCandles[i].Signal = 1
		} else if sellCondition {
			appliedCandles[i].Signal = -1
		} else {
			appliedCandles[i].Signal = 0
		}

		if i > 0 {
			appliedCandles[i].Position = appliedCandles[i].Signal - appliedCandles[i-1].Signal
		}

	}

	return appliedCandles
}

func max(a, b, c int) int {
	if a > b {
		if a > c {
			return a
		}
		return c
	}
	if b > c {
		return b
	}
	return c
}
