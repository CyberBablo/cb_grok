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
	rsiPeriod := params.RSIPeriod
	if rsiPeriod <= 0 {
		rsiPeriod = 14 // Default to 14 if invalid
	}
	
	atrPeriod := params.ATRPeriod
	if atrPeriod <= 0 {
		atrPeriod = 14 // Default to 14 if invalid
	}
	
	adxPeriod := params.ADXPeriod
	if adxPeriod <= 0 {
		adxPeriod = 14 // Default to 14 if invalid
	}
	
	maShortPeriod := params.MAShortPeriod
	if maShortPeriod <= 0 {
		maShortPeriod = 10 // Default to 10 if invalid
	}
	
	maLongPeriod := params.MALongPeriod
	if maLongPeriod <= 0 {
		maLongPeriod = 50 // Default to 50 if invalid
	}
	
	if len(candles) < max(maLongPeriod, params.EMALongPeriod, params.MACDLongPeriod) {
		return nil
	}

	rsi := indicators.CalculateRSI(candles, rsiPeriod)
	atr := indicators.CalculateATR(candles, atrPeriod)
	adx := indicators.CalculateADX(candles, adxPeriod)
	
	shortMA := indicators.CalculateSMA(candles, maShortPeriod)
	longMA := indicators.CalculateSMA(candles, maLongPeriod)
	
	macdFast := 12
	macdSlow := 26
	macdSignal := 9
	macd, macdSignalLine := indicators.CalculateMACD(candles, macdFast, macdSlow, macdSignal)
	
	bbPeriod := 20
	bbDeviation := 2.0
	upperBB, middleBB, lowerBB := indicators.CalculateBollingerBands(candles, bbPeriod, bbDeviation)
	
	stochK, stochD := indicators.CalculateStochastic(candles, 14, 3)
	
	lookbackPeriod := 20
	regime := DetectMarketRegime(candles, lookbackPeriod)

	var appliedCandles []models.AppliedOHLCV
	for i := range candles {
		bbWidth := 0.0
		if i >= bbPeriod && middleBB[i] > 0 {
			bbWidth = (upperBB[i] - lowerBB[i]) / middleBB[i]
		}
		
		var isTrending bool
		if i >= adxPeriod && i >= maLongPeriod {
			isTrending = adx[i] > params.ADXThreshold
		}
		
		isVolatile := bbWidth > 0.05 // 5% of price
		
		macdValue := 0.0
		macdSignalValue := 0.0
		if i < len(macd) && i < len(macdSignalLine) {
			macdValue = macd[i]
			macdSignalValue = macdSignalLine[i]
		}
		
		appliedCandles = append(appliedCandles, models.AppliedOHLCV{
			OHLCV:      candles[i],
			ATR:        atr[i],
			RSI:        rsi[i],
			ADX:        adx[i],
			ShortMA:    shortMA[i],
			LongMA:     longMA[i],
			ShortEMA:   upperBB[i],  // Using ShortEMA to store upperBB
			LongEMA:    lowerBB[i],  // Using LongEMA to store lowerBB
			MACD:       macdValue,
			MACDSignal: macdSignalValue,
			Trend:      isTrending,
			Volatility: isVolatile,
		})
	}

	for i := 1; i < len(appliedCandles); i++ {
		if i < maLongPeriod || i < bbPeriod {
			continue
		}
		
		var buySignal, sellSignal bool
		
		currentPrice := candles[i].Close
		previousPrice := candles[i-1].Close
		
		trendUp := shortMA[i] > longMA[i]
		trendDown := shortMA[i] < longMA[i]
		
		rsiOversold := rsi[i] < params.BuyRSIThreshold
		rsiOverbought := rsi[i] > params.SellRSIThreshold
		
		stochOversold := i < len(stochK) && i < len(stochD) && stochK[i] < 20 && stochD[i] < 20
		stochOverbought := i < len(stochK) && i < len(stochD) && stochK[i] > 80 && stochD[i] > 80
		stochCrossUp := i < len(stochK) && i < len(stochD) && 
			i-1 < len(stochK) && i-1 < len(stochD) && 
			stochK[i] > stochD[i] && stochK[i-1] <= stochD[i-1]
		stochCrossDown := i < len(stochK) && i < len(stochD) && 
			i-1 < len(stochK) && i-1 < len(stochD) && 
			stochK[i] < stochD[i] && stochK[i-1] >= stochD[i-1]
		
		priceRising := currentPrice > previousPrice
		priceFalling := currentPrice < previousPrice
		
		priceAboveUpper := currentPrice > upperBB[i]
		priceBelowLower := currentPrice < lowerBB[i]
		
		macdCrossUp := i < len(macd) && i < len(macdSignalLine) && 
			i-1 < len(macd) && i-1 < len(macdSignalLine) && 
			macd[i] > macdSignalLine[i] && macd[i-1] <= macdSignalLine[i-1]
		macdCrossDown := i < len(macd) && i < len(macdSignalLine) && 
			i-1 < len(macd) && i-1 < len(macdSignalLine) && 
			macd[i] < macdSignalLine[i] && macd[i-1] >= macdSignalLine[i-1]
		
		
		strongUptrend := trendUp && adx[i] > 25 && 
			i < len(macd) && i < len(macdSignalLine) && 
			macd[i] > 0 && macd[i] > macdSignalLine[i]
		strongDowntrend := trendDown && adx[i] > 25 && 
			i < len(macd) && i < len(macdSignalLine) && 
			macd[i] < 0 && macd[i] < macdSignalLine[i]
		
		pullbackInUptrend := strongUptrend && (priceBelowLower || (rsiOversold && priceRising))
		
		rallyInDowntrend := strongDowntrend && (priceAboveUpper || (rsiOverbought && priceFalling))
		
		bbSqueeze := i < len(upperBB) && i < len(lowerBB) && i < len(middleBB) && 
			middleBB[i] > 0 && (upperBB[i] - lowerBB[i]) / middleBB[i] < 0.03
		breakoutUp := bbSqueeze && stochCrossUp && macdCrossUp && priceRising
		breakoutDown := bbSqueeze && stochCrossDown && macdCrossDown && priceFalling
		
		reversalUp := priceBelowLower && rsiOversold && stochOversold && macdCrossUp && priceRising
		reversalDown := priceAboveUpper && rsiOverbought && stochOverbought && macdCrossDown && priceFalling
		
		switch regime {
		case TrendingMarket:
			if pullbackInUptrend {
				buySignal = true
			} else if rallyInDowntrend {
				sellSignal = true
			}
			
		case VolatileMarket:
			if reversalUp {
				buySignal = true
			} else if reversalDown {
				sellSignal = true
			}
			
		case RangeMarket:
			if breakoutUp || reversalUp {
				buySignal = true
			} else if breakoutDown || reversalDown {
				sellSignal = true
			}
		}
		
		if buySignal {
			appliedCandles[i].Signal = 1
		} else if sellSignal {
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

	adx := indicators.CalculateADX(recentCandles, 14) // Shorter ADX period for faster response
	atr := indicators.CalculateATR(recentCandles, 14) // Shorter ATR period for faster response
	stochK, stochD := indicators.CalculateStochastic(recentCandles, 14, 3)
	upperBB, middleBB, lowerBB := indicators.CalculateBollingerBands(recentCandles, 20, 2.0)
	
	latestADX := adx[len(adx)-1]
	latestStochK := stochK[len(stochK)-1]
	latestStochD := stochD[len(stochD)-1]
	
	bbWidth := 0.0
	if len(middleBB) > 0 && middleBB[len(middleBB)-1] > 0 {
		bbWidth = (upperBB[len(upperBB)-1] - lowerBB[len(lowerBB)-1]) / middleBB[len(middleBB)-1]
	}

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
	
	directionConsistency := 0
	for i := 1; i < len(recentCandles); i++ {
		if recentCandles[i].Close > recentCandles[i-1].Close {
			directionConsistency++
		} else if recentCandles[i].Close < recentCandles[i-1].Close {
			directionConsistency--
		}
	}
	directionConsistencyAbs := math.Abs(float64(directionConsistency)) / float64(len(recentCandles)-1)
	
	stochConsistency := 0
	stochCount := 0
	for i := 1; i < len(stochK); i++ {
		if stochK[i] > stochK[i-1] {
			stochConsistency++
		} else if stochK[i] < stochK[i-1] {
			stochConsistency--
		}
		stochCount++
	}
	stochConsistencyAbs := 0.0
	if stochCount > 0 {
		stochConsistencyAbs = math.Abs(float64(stochConsistency)) / float64(stochCount)
	}
	
	adxThreshold := 20.0  // ADX above 20 indicates trend
	atrThreshold := 0.006 // 0.6% normalized for volatility detection
	consistencyThreshold := 0.65 // 65% consistency indicates strong trend
	bbWidthThreshold := 0.05 // 5% width indicates volatility
	stochConsistencyThreshold := 0.7 // 70% consistency in stochastic movement

	isTrending := latestADX > adxThreshold || directionConsistencyAbs > consistencyThreshold
	isVolatile := avgATRPercentage > atrThreshold || bbWidth > bbWidthThreshold
	hasStrongMomentum := stochConsistencyAbs > stochConsistencyThreshold
	
	isOverbought := latestStochK > 80 && latestStochD > 80
	isOversold := latestStochK < 20 && latestStochD < 20
	
	if isTrending && hasStrongMomentum {
		return TrendingMarket // Strong trend with momentum confirmation
	} else if isVolatile && (isOverbought || isOversold) {
		return VolatileMarket // High volatility with extreme stochastic readings
	} else if isVolatile {
		return VolatileMarket // High volatility but no clear trend
	} else if bbWidth < 0.03 {
		return RangeMarket // Tight Bollinger Bands indicate range-bound market
	} else {
		return RangeMarket // Default to range market
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
