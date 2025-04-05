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
			
			isVolatile := false
			isTrending := false
			
			if i >= 20 && candles[i].ADX > 20.0 {
				isTrending = true
			}
			
			volatilityFactor := 1.0
			if i >= 5 {
				recentATR := 0.0
				for j := 0; j < 5; j++ {
					recentATR += candles[i-j].ATR
				}
				recentATR /= 5.0
				
				if candles[i].ATR > recentATR * 1.2 {
					isVolatile = true
					volatilityFactor = 0.7  // Reduce risk in volatile markets
				} else if candles[i].ATR < recentATR * 0.8 {
					volatilityFactor = 1.3  // Increase risk in low volatility
				}
			}
			
			var riskPercentage float64
			if isTrending {
				riskPercentage = 0.03  // Higher risk in trending markets
			} else if isVolatile {
				riskPercentage = 0.02  // Lower risk in volatile markets
			} else {
				riskPercentage = 0.025 // Default risk
			}
			
			riskAmount := capital * riskPercentage * volatilityFactor
			stopLossDistance := atr * params.StopLossMultiplier
			
			var positionSize float64
			if stopLossDistance > 0 {
				positionSize = riskAmount / stopLossDistance
			} else {
				positionSize = capital / buyPrice * 0.5 // Default to 50% if can't calculate
			}
			
			maxPositionSize := capital / buyPrice * (1 - b.Commission)
			if positionSize > maxPositionSize * 0.8 {
				positionSize = maxPositionSize * 0.8 // Cap at 80% of available capital
			} else if positionSize < maxPositionSize * 0.15 {
				positionSize = maxPositionSize * 0.15 // Minimum 15% of available capital
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
