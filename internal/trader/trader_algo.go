package trader

import (
	"cb_grok/pkg/models"
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

	var currentSignal int

	buyDelayCandles := 0
	for _, c := range appliedOHLCV[len(appliedOHLCV)-1-buyDelayCandles:] {
		if c.Signal == 1 {
			currentSignal = 1
			break
		}
	}

	currentPrice := currentCandle.Close

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

			t.state.cash += transactionAmount * currentPrice
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

			t.state.cash += transactionAmount * currentPrice
			t.state.assets = 0.0
			t.state.isPositionOpen = false

		} else if currentSignal == -1 && t.state.assets > 0 { // sell signal
			decision = DecisionSell

			transactionAmount = t.state.assets

			err := t.exch.CreateOrder(t.model.Symbol, "sell", transactionAmount, 0, 0)
			if err != nil {
				t.log.Error("create order failed", zap.Error(err))
			}

			t.state.cash += transactionAmount * currentPrice
			t.state.assets = 0.0
			t.state.isPositionOpen = false
		}
	} else if currentSignal == 1 && t.state.cash > 0 { // buy signal
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

	if action.Decision != DecisionHold {
		t.state.events = append(t.state.events, action)
	}

	return &action, nil
}
