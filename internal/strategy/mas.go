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

	shortMA := indicators.CalculateSMA(candles, params.MAShortPeriod)
	longMA := indicators.CalculateSMA(candles, params.MALongPeriod)
	rsi := indicators.CalculateRSI(candles, params.RSIPeriod)
	atr := indicators.CalculateATR(candles, params.ATRPeriod)
	emaShort := indicators.CalculateEMA(candles, params.EMAShortPeriod)
	emaLong := indicators.CalculateEMA(candles, params.EMALongPeriod)
	adx := indicators.CalculateADX(candles, params.ADXPeriod)
	macd, macdSignal := indicators.CalculateMACD(candles, params.MACDShortPeriod, params.MACDLongPeriod, params.MACDSignalPeriod)

	lookbackPeriod := 20
	regime := DetectMarketRegime(candles, lookbackPeriod)

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
			ADX:        adx[i],
			MACD:       macd[i],
			MACDSignal: macdSignal[i],
		})
	}

	for i := 1; i < len(appliedCandles); i++ {
		var buyCondition, sellCondition bool

		maCrossover := shortMA[i] > longMA[i] && shortMA[i-1] <= longMA[i-1]
		maCrossunder := shortMA[i] < longMA[i] && shortMA[i-1] >= longMA[i-1]
		
		macdCrossover := macd[i] > macdSignal[i] && macd[i-1] <= macdSignal[i-1]
		macdCrossunder := macd[i] < macdSignal[i] && macd[i-1] >= macdSignal[i-1]
		
		rsiOversold := rsi[i] < 30
		rsiOverbought := rsi[i] > 70

		switch regime {
		case TrendingMarket:
			if adx[i] > 20 { // Strong trend
				buyCondition = (maCrossover || macdCrossover) && trend[i]
				sellCondition = (maCrossunder || macdCrossunder) && !trend[i]
			} else { // Weak trend
				buyCondition = (maCrossover || (rsiOversold && trend[i]))
				sellCondition = (maCrossunder || (rsiOverbought && !trend[i]))
			}
		
		case VolatileMarket:
			buyCondition = rsiOversold || macdCrossover
			sellCondition = rsiOverbought || macdCrossunder
		
		case RangeMarket:
			buyCondition = rsi[i] < 40 && rsi[i] > rsi[i-1]
			sellCondition = rsi[i] > 60 && rsi[i] < rsi[i-1]
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
