package trader

import (
	"cb_grok/pkg/models"
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

	var currentSignal int

	for i, _ := range appliedOHLCV {
		if appliedOHLCV[len(appliedOHLCV)-i-1].Signal == 1 {
			fmt.Println("купил?", appliedOHLCV[len(appliedOHLCV)-i-1].Signal, len(appliedOHLCV)-i-1, len(appliedOHLCV))
			if len(appliedOHLCV)-(len(appliedOHLCV)-i-1) < 3 {
				currentSignal = 1
				fmt.Println("есть сигнал по покупке", appliedOHLCV[len(appliedOHLCV)-i-1].Signal, len(appliedOHLCV)-i-1, len(appliedOHLCV))
				//os.Exit(0)
			}
			//os.Exit(0)
			break
		} else if appliedOHLCV[len(appliedOHLCV)-i-1].Signal == -1 {
			fmt.Println("продал?", appliedOHLCV[len(appliedOHLCV)-i-1].Signal, len(appliedOHLCV)-i-1, len(appliedOHLCV))

			if len(appliedOHLCV)-(len(appliedOHLCV)-i-1) < 1 {
				currentSignal = -1
				fmt.Println("есть сигнал о продаже", appliedOHLCV[len(appliedOHLCV)-i-1].Signal, len(appliedOHLCV)-i-1, len(appliedOHLCV))
				//os.Exit(0)
			}
			break

		}
	}

	//currentSignal := currentCandle.Signal
	currentPrice := currentCandle.Close

	//b, _ := json.Marshal(currentCandle)
	//fmt.Printf("<<<<< %s\n", string(b))

	atr := currentCandle.ATR

	var (
		decision        = DecisionHold
		decisionTrigger = TriggerSignal
	)

	transactionAmount := 0.0
	//if currentSignal == 1 {
	//	fmt.Println(t.state.cash)
	//	os.Exit(0)
	//}

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
		fmt.Println("DESICION BUY?")
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
