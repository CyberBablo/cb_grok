package strategy

import (
	"cb_grok/internal/indicators"
	"cb_grok/pkg/models"
	"math"
)

type MovingAverageStrategy struct{}

func NewMovingAverageStrategy() Strategy {
	return &MovingAverageStrategy{}
}

func (s *MovingAverageStrategy) Apply(candles []models.OHLCV, params StrategyParams) []models.AppliedOHLCV {
	if len(candles) < max(params.MALongPeriod, params.EMALongPeriod, params.MACDLongPeriod) {
		return nil
	}

	lookbackPeriod := 20
	regime := DetectMarketRegime(candles, lookbackPeriod)
	
	adjustedParams := params
	switch regime {
	case TrendingMarket:
		adjustedParams.UseTrendFilter = true
		adjustedParams.TakeProfitMultiplier *= 1.2
	case VolatileMarket:
		adjustedParams.StopLossMultiplier *= 0.8
		adjustedParams.TakeProfitMultiplier *= 0.8
		adjustedParams.UseRSIFilter = true
	case RangeMarket:
		adjustedParams.UseRSIFilter = true
		adjustedParams.TakeProfitMultiplier *= 0.9
	}

	shortMA := indicators.CalculateSMA(candles, adjustedParams.MAShortPeriod)
	longMA := indicators.CalculateSMA(candles, adjustedParams.MALongPeriod)
	rsi := indicators.CalculateRSI(candles, adjustedParams.RSIPeriod)
	atr := indicators.CalculateATR(candles, adjustedParams.ATRPeriod)
	emaShort := indicators.CalculateEMA(candles, adjustedParams.EMAShortPeriod)
	emaLong := indicators.CalculateEMA(candles, adjustedParams.EMALongPeriod)
	adx := indicators.CalculateADX(candles, adjustedParams.ADXPeriod)
	macd, macdSignal := indicators.CalculateMACD(candles, adjustedParams.MACDShortPeriod, adjustedParams.MACDLongPeriod, adjustedParams.MACDSignalPeriod)

	trend := make([]bool, len(candles))
	volatility := make([]bool, len(candles))
	for i := range candles {
		trend[i] = emaShort[i] > emaLong[i]
		volatility[i] = atr[i] > adjustedParams.ATRThreshold
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
			ADX:        adx[i],
			MACD:       macd[i],
			MACDSignal: macdSignal[i],
		})
	}

	for i := 1; i < len(appliedCandles); i++ {
		buyCondition := shortMA[i] > longMA[i] && macd[i] > macdSignal[i] && adx[i] > adjustedParams.ADXThreshold
		sellCondition := shortMA[i] < longMA[i] && macd[i] < macdSignal[i] && adx[i] > adjustedParams.ADXThreshold

		if adjustedParams.UseRSIFilter {
			buyCondition = buyCondition && rsi[i] < adjustedParams.BuyRSIThreshold
			sellCondition = sellCondition && rsi[i] > adjustedParams.SellRSIThreshold
		}

		if adjustedParams.UseTrendFilter {
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

type MarketRegime int

const (
	RangeMarket MarketRegime = iota
	TrendingMarket
	VolatileMarket
)

func DetectMarketRegime(candles []models.OHLCV, lookbackPeriod int) MarketRegime {
	if len(candles) < lookbackPeriod {
		return RangeMarket // Default to range market if not enough data
	}

	recentCandles := candles[len(candles)-lookbackPeriod:]
	
	returns := make([]float64, lookbackPeriod-1)
	for i := 1; i < lookbackPeriod; i++ {
		returns[i-1] = (recentCandles[i].Close - recentCandles[i-1].Close) / recentCandles[i-1].Close
	}
	
	meanReturn := 0.0
	for _, r := range returns {
		meanReturn += r
	}
	meanReturn /= float64(len(returns))
	
	variance := 0.0
	for _, r := range returns {
		variance += (r - meanReturn) * (r - meanReturn)
	}
	volatility := math.Sqrt(variance / float64(len(returns)))
	
	highestHigh := recentCandles[0].High
	lowestLow := recentCandles[0].Low
	for _, candle := range recentCandles {
		if candle.High > highestHigh {
			highestHigh = candle.High
		}
		if candle.Low < lowestLow {
			lowestLow = candle.Low
		}
	}
	
	priceRange := highestHigh - lowestLow
	startPrice := recentCandles[0].Close
	endPrice := recentCandles[len(recentCandles)-1].Close
	directionalMove := math.Abs(endPrice - startPrice)
	
	directionalStrength := 0.0
	if priceRange > 0 {
		directionalStrength = directionalMove / priceRange
	}
	
	volatilityThreshold := 0.015 // 1.5% daily volatility is considered high
	
	directionalThreshold := 0.6 // 60% of the range is directional
	
	if volatility > volatilityThreshold && directionalStrength > directionalThreshold {
		return TrendingMarket
	} else if volatility > volatilityThreshold {
		return VolatileMarket
	} else {
		return RangeMarket
	}
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
