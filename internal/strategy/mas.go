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
	adx := indicators.CalculateADX(candles, params.ADXPeriod) // Uncommented ADX calculation
	macd, macdSignal := indicators.CalculateMACD(candles, params.MACDShortPeriod, params.MACDLongPeriod, params.MACDSignalPeriod)

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
			ADX:        adx[i], // Uncommented ADX field
			MACD:       macd[i],
			MACDSignal: macdSignal[i],
		})
	}

	for i := 1; i < len(appliedCandles); i++ {
		buyScore := 0.0
		sellScore := 0.0
		
		if shortMA[i] > longMA[i] {
			buyScore += params.MAWeight
		} else if shortMA[i] < longMA[i] {
			sellScore += params.MAWeight
		}
		
		if macd[i] > macdSignal[i] {
			buyScore += params.MACDWeight
		} else if macd[i] < macdSignal[i] {
			sellScore += params.MACDWeight
		}
		
		if params.UseRSIFilter {
			if rsi[i] < params.BuyRSIThreshold {
				buyScore += params.RSIWeight
			} else if rsi[i] > params.SellRSIThreshold {
				sellScore += params.RSIWeight
			}
		}
		
		if params.UseADXFilter && adx[i] > params.ADXThreshold {
			buyScore += params.ADXWeight
			sellScore += params.ADXWeight
		}
		
		if params.UseTrendFilter {
			if trend[i] && volatility[i] {
				buyScore += params.TrendWeight
			} else if !trend[i] && volatility[i] {
				sellScore += params.TrendWeight
			}
		}
		
		if buyScore >= params.BuyThreshold {
			appliedCandles[i].Signal = 1
		} else if sellScore >= params.SellThreshold {
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
