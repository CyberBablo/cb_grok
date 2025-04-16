package trader

import (
	"cb_grok/pkg/models"
	"go.uber.org/zap"
	"strings"
)

func (t *trader) processAlgo(candle models.OHLCV) (*Action, error) {
	t.state.ohlcv = append(t.state.ohlcv, candle)

	appliedOHLCV := t.strategy.ApplyIndicators(t.state.ohlcv, t.model.StrategyParams)
	if appliedOHLCV == nil {
		return nil, nil
	}

	return t.algo(appliedOHLCV)
}

func (t *trader) BacktestAlgo(appliedOHLCV []models.AppliedOHLCV) (*Action, error) {
	tmpOHLCV := make([]models.OHLCV, len(appliedOHLCV))
	for i := range appliedOHLCV {
		tmpOHLCV[i] = appliedOHLCV[i].OHLCV
	}
	t.state.ohlcv = tmpOHLCV

	return t.algo(appliedOHLCV)
}

func (t *trader) algo(appliedOHLCV []models.AppliedOHLCV) (*Action, error) {
	appliedOHLCV = t.strategy.ApplySignals(appliedOHLCV, t.model.StrategyParams)
	if appliedOHLCV == nil {
		return nil, nil
	}

	t.state.appliedOHLCV = appliedOHLCV

	currentCandle := appliedOHLCV[len(appliedOHLCV)-1]

	//if len(appliedOHLCV) > 2870 {
	//	b, _ := json.Marshal(currentCandle)
	//	fmt.Printf("%s\n", string(b))
	//}

	currentSignal := currentCandle.Signal
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

		t.state.stopLoss = currentPrice - atr*t.settings.StopLossMultiplier
		t.state.takeProfit = currentPrice + atr*t.settings.TakeProfitMultiplier
	}

	portfolioValue := t.state.cash + t.state.assets*currentPrice
	t.state.portfolioValues = append(t.state.portfolioValues, PortfolioValue{
		Timestamp: currentCandle.Timestamp,
		Value:     portfolioValue,
	})

	action := Action{
		Timestamp:       currentCandle.Timestamp,
		Decision:        decision,
		DecisionTrigger: decisionTrigger,
		Price:           currentPrice,
		AssetAmount:     transactionAmount,
		AssetCurrency:   strings.Split(t.model.Symbol, "/")[0],
		Comment:         "",
		PortfolioValue:  portfolioValue,
	}

	if action.Decision != DecisionHold {
		t.state.orders = append(t.state.orders, action)
	}

	return &action, nil
}
