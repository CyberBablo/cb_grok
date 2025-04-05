package backtest

import (
	"cb_grok/internal/strategy"
	"cb_grok/pkg/models"
	"fmt"
	"math"
)

// Order представляет информацию о торговом ордере
type Order struct {
	Action    string  `json:"action"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
	Reason    string  `json:"reason,omitempty"`
}

// Backtest определяет интерфейс бэктеста
type Backtest interface {
	Run(candles []models.OHLCV, params strategy.StrategyParams) (sharpeRatio float64, orders []Order, finalCapital float64, maxDrawdown float64, winRate float64, err error)
}

// backtestImpl реализует логику бэктеста
type backtestImpl struct {
	InitialCapital       float64
	Commission           float64
	SlippagePercent      float64
	Spread               float64
	StopLossMultiplier   float64
	TakeProfitMultiplier float64
}

// NewBacktest создает новый экземпляр бэктеста
func NewBacktest() Backtest {
	return &backtestImpl{
		InitialCapital:       10000.0,
		Commission:           0.001,  // 0.1%
		SlippagePercent:      0.001,  // 0.1%
		Spread:               0.0002, // 0.02%
		StopLossMultiplier:   1.5,
		TakeProfitMultiplier: 3.0,
	}
}

// Run выполняет бэктест
func (b *backtestImpl) Run(ohlcv []models.OHLCV, params strategy.StrategyParams) (float64, []Order, float64, float64, float64, error) {
	str := strategy.NewMovingAverageStrategy()
	candles := str.Apply(ohlcv, params)
	if candles == nil {
		return 0, nil, b.InitialCapital, 0, 0, nil
	}

	for _, c := range candles {
		if c.ATR == 0 {
			return 0, nil, b.InitialCapital, 0, 0, fmt.Errorf("ATR is required for backtest")
		}
	}

	capital := b.InitialCapital
	position := 0.0
	entryPrice := 0.0
	stopLoss := 0.0
	takeProfit := 0.0
	var orders []Order
	portfolioValues := make([]float64, len(candles))

	for i := 1; i < len(candles)-1; i++ {
		signal := candles[i].Signal
		nextOpen := candles[i+1].Open
		timestamp := candles[i+1].Timestamp
		atr := candles[i].ATR

		if position > 0 {
			currentClose := candles[i].Close
			if currentClose <= stopLoss {
				sellPrice := nextOpen * (1 - b.SlippagePercent - b.Spread)
				capital = position * sellPrice * (1 - b.Commission)
				orders = append(orders, Order{Action: "sell", Amount: position, Price: sellPrice, Timestamp: timestamp, Reason: "stop_loss"})
				position = 0
			} else if currentClose >= takeProfit {
				sellPrice := nextOpen * (1 - b.SlippagePercent - b.Spread)
				capital = position * sellPrice * (1 - b.Commission)
				orders = append(orders, Order{Action: "sell", Amount: position, Price: sellPrice, Timestamp: timestamp, Reason: "take_profit"})
				position = 0
			}
		}

		if signal == 1 && capital > 0 {
			buyPrice := nextOpen * (1 + b.SlippagePercent + b.Spread)

			upperBB := candles[i].ShortMA
			lowerBB := candles[i].LongMA
			middleBB := candles[i].ShortEMA
			
			bbWidth := 0.0
			if middleBB > 0 {
				bbWidth = (upperBB - lowerBB) / middleBB
			}
			
			isVolatile := bbWidth > 0.05 // Wide bands (>5% of price)
			isTrending := candles[i].Trend // Using ADX-based trend detection
			bbSqueeze := bbWidth < 0.03    // Narrow bands (<3% of price)
			
			pricePosition := 0.0
			if upperBB > lowerBB {
				pricePosition = (candles[i].Close - lowerBB) / (upperBB - lowerBB) // 0 = at lower band, 1 = at upper band
			}
			
			var riskPercentage float64
			
			trendStrength := 1.0
			if candles[i].ADX > 30.0 {
				trendStrength = 1.5 // Very strong trend
			} else if candles[i].ADX > 20.0 {
				trendStrength = 1.2 // Strong trend
			} else if candles[i].ADX < 15.0 {
				trendStrength = 0.8 // Weak trend
			}
			
			if i >= 10 {
				priceChange := math.Abs((candles[i].Close - candles[i-10].Close) / candles[i-10].Close)
				if priceChange > 0.05 { // 5% change in 10 candles indicates strong movement
					trendStrength *= 1.2
				}
			}
			
			if isTrending {
				if pricePosition > 0.5 && candles[i].MACD > candles[i].MACDSignal { // Uptrend
					riskPercentage = 0.05 * trendStrength
				} else if pricePosition < 0.5 && candles[i].MACD < candles[i].MACDSignal { // Downtrend
					riskPercentage = 0.05 * trendStrength
				} else {
					riskPercentage = 0.03 * trendStrength // Counter-trend move
				}
			} else if isVolatile {
				if pricePosition < 0.2 { // Near lower band - potential buy
					riskPercentage = 0.04 // More aggressive on potential bottoms
				} else if pricePosition > 0.8 { // Near upper band - potential sell
					riskPercentage = 0.02 // More conservative near tops
				} else {
					riskPercentage = 0.03 // Middle of range
				}
			} else if bbSqueeze {
				riskPercentage = 0.04 * trendStrength // More aggressive for potential breakout
			} else {
			riskPercentage = 0.02 // Reduced base risk percentage
			}
			
			if len(orders) >= 6 {
				profitableTrades := 0
				totalProfit := 0.0
				for j := len(orders) - 6; j < len(orders); j += 2 {
					if j+1 < len(orders) && orders[j].Action == "buy" && orders[j+1].Action == "sell" {
						if orders[j+1].Price > orders[j].Price {
							profitableTrades++
							totalProfit += (orders[j+1].Price - orders[j].Price) / orders[j].Price
						} else {
							totalProfit -= (orders[j].Price - orders[j+1].Price) / orders[j].Price
						}
					}
				}
				
				if profitableTrades >= 3 && totalProfit > 0.05 {
					riskPercentage *= 1.3 // Significant increase for strong performance
				} else if profitableTrades >= 2 {
					riskPercentage *= 1.1 // Modest increase for decent performance
				} else if profitableTrades <= 1 {
					riskPercentage *= 0.7 // Significant decrease for poor performance
				}
			}
			
			volatilityFactor := 1.0
			if bbWidth > 0.06 {
				volatilityFactor = 0.6 // More conservative in very volatile markets
			} else if bbWidth > 0.04 {
				volatilityFactor = 0.8 // Moderately conservative in volatile markets
			} else if bbWidth < 0.02 {
				volatilityFactor = 1.2 // Slightly more aggressive in low volatility
			}
			
			if riskPercentage > 0.04 {
				riskPercentage = 0.04
			}
			
			riskAmount := capital * riskPercentage * volatilityFactor
			
			var stopLossDistance float64
			if isTrending {
				stopLossDistance = atr * params.StopLossMultiplier * 1.2
			} else if isVolatile {
				stopLossDistance = atr * params.StopLossMultiplier * 0.8
			} else if bbSqueeze {
				stopLossDistance = atr * params.StopLossMultiplier * 0.7 // Tighter stops in squeeze conditions
			} else {
				stopLossDistance = atr * params.StopLossMultiplier
			}
			
			var positionSize float64
			if stopLossDistance > 0 {
				positionSize = riskAmount / stopLossDistance
			} else {
				positionSize = capital / buyPrice * 0.5 // Default to 50% if can't calculate
			}
			
			maxPositionSize := capital / buyPrice * (1 - b.Commission)
			if positionSize > maxPositionSize * 0.95 {
				positionSize = maxPositionSize * 0.95 // Cap at 95% of available capital for more aggressive sizing
			} else if positionSize < maxPositionSize * 0.15 {
				positionSize = maxPositionSize * 0.15 // Minimum 15% of available capital for more meaningful positions
			}

			position = positionSize
			capital = capital - (position * buyPrice)
			entryPrice = buyPrice
			stopLoss = entryPrice - stopLossDistance
			takeProfit = entryPrice + atr*params.TakeProfitMultiplier
			orders = append(orders, Order{Action: "buy", Amount: position, Price: buyPrice, Timestamp: timestamp})
		} else if signal == -1 && position > 0 {
			sellPrice := nextOpen * (1 - b.SlippagePercent - b.Spread)
			capital = position * sellPrice * (1 - b.Commission)
			orders = append(orders, Order{Action: "sell", Amount: position, Price: sellPrice, Timestamp: timestamp, Reason: "signal"})
			position = 0
		}

		portfolioValues[i] = capital + position*candles[i+1].Close
	}

	if position > 0 {
		lastPrice := candles[len(candles)-1].Close * (1 - b.SlippagePercent - b.Spread)
		capital = position * lastPrice * (1 - b.Commission)
		orders = append(orders, Order{Action: "sell", Amount: position, Price: lastPrice, Timestamp: candles[len(candles)-1].Timestamp, Reason: "end_of_backtest"})
	}

	portfolioValues[len(candles)-1] = capital

	if len(portfolioValues) > 1 {
		returns := make([]float64, len(portfolioValues)-1)
		for i := 1; i < len(portfolioValues); i++ {
			if portfolioValues[i-1] != 0 {
				returns[i-1] = (portfolioValues[i] - portfolioValues[i-1]) / portfolioValues[i-1]
			}
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
		stdDev := math.Sqrt(variance / float64(len(returns)))

		sharpeRatio := 0.0
		if stdDev != 0 {
			sharpeRatio = meanReturn / stdDev * math.Sqrt(252) // Годовой Sharpe Ratio
		}

		// Расчет Max Drawdown
		maxDrawdown := 0.0
		peak := portfolioValues[0]
		for _, value := range portfolioValues {
			if value > peak {
				peak = value
			}
			drawdown := (peak - value) / peak * 100
			if drawdown > maxDrawdown {
				maxDrawdown = drawdown
			}
		}

		// Расчет Win Rate
		var wins, totalTrades int
		for i := 0; i < len(orders)-1; i += 2 {
			if orders[i].Action == "buy" && orders[i+1].Action == "sell" {
				if orders[i+1].Price > orders[i].Price {
					wins++
				}
				totalTrades++
			}
		}
		winRate := 0.0
		if totalTrades > 0 {
			winRate = float64(wins) / float64(totalTrades) * 100
		}

		return sharpeRatio, orders, capital, maxDrawdown, winRate, nil
	}

	return 0, orders, capital, 0, 0, nil
}
