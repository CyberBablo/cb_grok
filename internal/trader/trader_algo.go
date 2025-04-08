package trader

import (
	"cb_grok/pkg/models"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"strings"
)

func (t *trader) algo(candle models.OHLCV) (*Action, error) {
	t.state.ohlcv = append(t.state.ohlcv, candle)

	appliedOHLCV := t.strategy.Apply(t.state.ohlcv, t.model.StrategyParams)
	if appliedOHLCV == nil {
		return nil, nil
	}

	currentCandle := appliedOHLCV[len(appliedOHLCV)-1]

	currentSignal := currentCandle.Signal
	currentPrice := currentCandle.Close

	b, _ := json.Marshal(currentCandle)
	fmt.Printf("<<<<< %s\n", string(b))

	atr := currentCandle.ATR

	var (
		decision        = DecisionHold
		decisionTrigger = TriggerSignal
	)

	transactionAmount := 0.0

	if t.state.isPositionOpen {
		if currentPrice <= t.state.stopLoss { // sell stop-loss
			decision = DecisionSell
			decisionTrigger = TriggerStopLoss

			transactionAmount = t.state.assets

			err := t.exch.CreateOrder(t.model.Symbol, "sell", transactionAmount, 0, 0)
			if err != nil {
				t.log.Error("create order failed", zap.Error(err))
			}

			t.state.cash += t.state.assets * currentPrice
			t.state.assets = 0.0
			t.state.isPositionOpen = false

		} else if currentPrice >= t.state.takeProfit { // sell take-profit
			decision = DecisionSell
			decisionTrigger = TriggerTakeProfit

			transactionAmount = t.state.assets

			err := t.exch.CreateOrder(t.model.Symbol, "sell", transactionAmount, 0, 0)
			if err != nil {
				t.log.Error("create order failed", zap.Error(err))
			}

			t.state.cash += t.state.assets * currentPrice
			t.state.assets = 0.0
			t.state.isPositionOpen = false

		} else if currentSignal == -1 { // sell signal
			decision = DecisionSell

			transactionAmount = t.state.assets

			err := t.exch.CreateOrder(t.model.Symbol, "sell", transactionAmount, 0, 0)
			if err != nil {
				t.log.Error("create order failed", zap.Error(err))
			}

			t.state.cash += t.state.assets * currentPrice
			t.state.assets = 0.0
			t.state.isPositionOpen = false
		}
	} else if currentSignal == 1 && t.state.cash > 0 {
		decision = DecisionBuy

		transactionAmount = t.state.cash / currentPrice

		err := t.exch.CreateOrder(t.model.Symbol, "buy", transactionAmount, 0, 0)
		if err != nil {
			t.log.Error("create order failed", zap.Error(err))
		}

		t.state.cash = 0.0
		t.state.assets = transactionAmount
		t.state.isPositionOpen = true

		t.state.entryPrice = currentPrice
		t.state.stopLoss = currentPrice - atr*t.settings.StopLossMultiplier
		t.state.takeProfit = currentPrice + atr*t.settings.TakeProfitMultiplier
	}

	portfolioValue := t.state.cash + t.state.assets*currentPrice

	action := Action{
		Timestamp:       currentCandle.Timestamp,
		Decision:        decision,
		DecisionTrigger: decisionTrigger,
		AssetAmount:     transactionAmount,
		AssetCurrency:   strings.Split(t.model.Symbol, "/")[0],
		Comment:         "",
		PortfolioValue:  portfolioValue,
	}
	//
	//t.log.Info("trade algo action",
	//	zap.Int64("timeframe", action.Timestamp),
	//	zap.String("decision", string(action.Decision)),
	//	zap.String("trigger", string(action.DecisionTrigger)),
	//	zap.Float64("asset_amount", action.AssetAmount),
	//	zap.String("asset_curr", action.AssetCurrency),
	//	zap.Float64("portfolio_usdt", action.PortfolioValue),
	//)

	if action.Decision != DecisionHold {
		t.state.events = append(t.state.events, action)
	}

	return &action, nil
}
