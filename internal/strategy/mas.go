package strategy

import (
	"cb_grok/internal/indicators"
	"cb_grok/pkg/models"
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
		adjustedParams.TakeProfitMultiplier *= 1.1 // Slightly increase take profit
		adjustedParams.ADXThreshold *= 0.9         // Slightly lower ADX threshold
	case VolatileMarket:
		adjustedParams.StopLossMultiplier *= 0.9 // Tighter stop loss
		adjustedParams.UseRSIFilter = true       // Enable RSI filter
	case RangeMarket:
		adjustedParams.UseRSIFilter = true // Enable RSI filter
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
		buyCondition := shortMA[i] > longMA[i] && macd[i] > macdSignal[i]
		sellCondition := shortMA[i] < longMA[i] && macd[i] < macdSignal[i]

		if adx[i] > 25.0 && trend[i] {
			buyCondition = buyCondition && true
		} else if adx[i] > 25.0 && !trend[i] {
			sellCondition = sellCondition && true
		}

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

	adx := indicators.CalculateADX(recentCandles, 20) // Use optimized ADX period

	atr := indicators.CalculateATR(recentCandles, 16) // Use optimized ATR period

	latestADX := adx[len(adx)-1]

	avgATRPercentage := 0.0
	count := 0
	for i := len(recentCandles) - 5; i < len(recentCandles); i++ {
		if i >= 0 && i < len(atr) && i < len(recentCandles) {
			avgATRPercentage += atr[i] / recentCandles[i].Close
			count++
		}
	}
	if count > 0 {
		avgATRPercentage /= float64(count)
	}

	adxThreshold := 20.0  // Lower threshold to detect more trends
	atrThreshold := 0.005 // 0.5% normalized for better volatility detection

	if latestADX > adxThreshold {
		return TrendingMarket // Strong trend detected
	} else if avgATRPercentage > atrThreshold {
		return VolatileMarket // High volatility but no clear trend
	} else {
		return RangeMarket // Low volatility, no clear trend
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
