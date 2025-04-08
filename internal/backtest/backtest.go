package backtest

import (
	"cb_grok/internal/strategy"
	"cb_grok/pkg/models"
	"fmt"
	"math"
)

type Backtest interface {
	Run(candles []models.OHLCV, params strategy.StrategyParams) (*BacktestResult, error)
}

type backtest struct {
	InitialCapital       float64
	Commission           float64
	SlippagePercent      float64
	Spread               float64
	StopLossMultiplier   float64
	TakeProfitMultiplier float64
}

func NewBacktest() Backtest {
	return &backtest{
		InitialCapital:       10000.0,
		Commission:           0.001,  // 0.1%
		SlippagePercent:      0.001,  // 0.1%
		Spread:               0.0002, // 0.02%
		StopLossMultiplier:   5,
		TakeProfitMultiplier: 5,
	}
}

func (b *backtest) Run(ohlcv []models.OHLCV, params strategy.StrategyParams) (*BacktestResult, error) {
	str := strategy.NewLinearBiasStrategy()
	candles := str.Apply(ohlcv, params)
	if candles == nil {
		return nil, fmt.Errorf("no candles after strategy apply")
	}

	for _, c := range candles {
		if c.ATR == 0 {
			return nil, fmt.Errorf("ATR is required for backtest")
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
			position = capital / buyPrice * (1 - b.Commission)
			capital = 0
			entryPrice = buyPrice
			stopLoss = entryPrice - atr*b.StopLossMultiplier
			takeProfit = entryPrice + atr*b.TakeProfitMultiplier
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

		return &BacktestResult{
			SharpeRatio:  sharpeRatio,
			Orders:       orders,
			FinalCapital: capital,
			MaxDrawdown:  maxDrawdown,
			WinRate:      winRate,
		}, nil

	}

	return &BacktestResult{
		Orders:       orders,
		FinalCapital: capital,
	}, nil
}
