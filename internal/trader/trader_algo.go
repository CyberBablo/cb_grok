package trader

import (
	"cb_grok/pkg/models"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"strings"
)

func (t *trader) processAlgo(candle models.OHLCV) (*Action, error) {
	if len(t.state.ohlcv) > 0 && t.state.ohlcv[len(t.state.ohlcv)-1].Timestamp == candle.Timestamp {
		t.state.ohlcv[len(t.state.ohlcv)-1] = candle
	} else {
		t.state.ohlcv = append(t.state.ohlcv, candle)
	}

	candleLog, _ := json.Marshal(candle)
	t.log.Info("trader: new candle has been processed", zap.String("candle", string(candleLog)))

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

	currentSignal := currentCandle.Signal
	currentPrice := currentCandle.Close

	atr := currentCandle.ATR

	var (
		decision        = DecisionHold
		decisionTrigger = TriggerSignal
	)

	transactionAmount := 0.0

	orders, err := t.orderUC.GetActiveOrders(context.Background())
	if err != nil {
		t.log.Error("failed to fetch active orders", zap.Error(err))
		return nil, err
	}

	isPositionOpen := len(orders) > 0

	if isPositionOpen {
		if currentSignal == -1 { // sell signal
			decision = DecisionSell

			transactionAmount = t.state.assets

			err := t.orderUC.CreateSpotMarketOrder(t.model.Symbol, "sell", transactionAmount)
			if err != nil {
				t.log.Error("create order failed", zap.Error(err))
			}
		}
	} else {
		if currentSignal == 1 { // buy signal
			decision = DecisionBuy

			transactionAmount = t.state.cash / currentPrice

			err := t.orderUC.CreateSpotMarketOrder(t.model.Symbol, "buy", transactionAmount)
			if err != nil {
				t.log.Error("create order failed", zap.Error(err))
			}
		}
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
