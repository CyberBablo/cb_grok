package trader

import (
	order_model "cb_grok/internal/order/model"
	"cb_grok/pkg/models"
	"context"
	"encoding/json"
	"fmt"
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
	t.log.Info("trader: new candle has been processed", zap.Int("total_length", len(t.state.ohlcv)), zap.String("candle", string(candleLog)))

	appliedOHLCV := t.strategy.ApplyIndicators(t.state.ohlcv, t.model.StrategyParams)
	if appliedOHLCV == nil {
		t.log.Info("trader: not enough candles in the dataset")
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

	t.log.Info("trader: processed signal", zap.Int("sig", currentSignal))

	var (
		decision        = DecisionHold
		decisionTrigger = TriggerSignal
	)

	transactionAmount := 0.0

	symbol, err := t.orderUC.GetSymbolByCode(t.model.Symbol)
	if err != nil {
		t.log.Error("failed to get symbol by code", zap.Error(err))
		return nil, err
	}

	orders, err := t.orderUC.GetActiveOrders(context.Background())
	if err != nil {
		t.log.Error("failed to fetch active orders", zap.Error(err))
		return nil, err
	}

	if len(orders) > 1 {
		t.log.Error("unexpected situation: there are more than 1 opened orders")
		return nil, fmt.Errorf("unexpected situation: there are more than 1 opened orders")
	}

	lastOrder, err := t.orderUC.GetLastOrder()
	if err != nil {
		t.log.Error("failed to fetch last order", zap.Error(err))
		return nil, err
	}

	allowSell := lastOrder != nil && lastOrder.SideID == int(order_model.OrderSideBuy) && lastOrder.StatusID == int(order_model.OrderStatusFilled)
	allowBuy := lastOrder == nil || (lastOrder.SideID == int(order_model.OrderSideSell) && lastOrder.StatusID == int(order_model.OrderStatusFilled))

	if allowBuy && allowSell {
		t.log.Error("trade_algo: unexpected situation: allowed both buy and sell")
		return nil, fmt.Errorf("trade_algo: unexpected situation: allowed both buy and sell")
	}

	if allowSell {
		if currentSignal == -1 { // sell signal
			decision = DecisionSell

			transactionAmount = *lastOrder.QuoteQty

			err = t.orderUC.CreateSpotMarketOrder(t.model.Symbol, "sell", transactionAmount, nil, nil)
			if err != nil {
				t.log.Error("create order failed", zap.Error(err))
			}
		} else if currentPrice >= *lastOrder.TakeProfitPrice { // sell take-profit
			decision = DecisionSell
			decisionTrigger = TriggerTakeProfit

			transactionAmount = *lastOrder.QuoteQty

			err = t.orderUC.CreateSpotMarketOrder(t.model.Symbol, "sell", transactionAmount, nil, nil)
			if err != nil {
				t.log.Error("create order failed", zap.Error(err))
			}

		} else if currentPrice <= *lastOrder.StopLossPrice {
			decision = DecisionSell
			decisionTrigger = TriggerStopLoss

			transactionAmount = *lastOrder.QuoteQty

			err = t.orderUC.CreateSpotMarketOrder(t.model.Symbol, "sell", transactionAmount, nil, nil)
			if err != nil {
				t.log.Error("create order failed", zap.Error(err))
			}
		}

	} else if allowBuy {
		if currentSignal == 1 { // buy signal
			bal, err := t.exch.GetAvailableSpotWalletBalance(symbol.Quote)
			if err != nil {
				t.log.Error("failed to get available wallet balance", zap.Error(err))
				return nil, err
			}

			decision = DecisionBuy

			transactionAmount = min(bal, 100)

			stopLoss := currentPrice - currentCandle.ATR*t.settings.StopLossMultiplier
			takeProfit := currentPrice + currentCandle.ATR*t.settings.TakeProfitMultiplier

			err = t.orderUC.CreateSpotMarketOrder(t.model.Symbol, "buy", transactionAmount, &takeProfit, &stopLoss)
			if err != nil {
				t.log.Error("create order failed", zap.Error(err))
			}
		}
	}

	portfolioValue := 0.0
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
		// Calculate profit for sell orders
		if action.Decision == DecisionSell && len(t.state.orders) > 0 {
			// Find the last buy order
			for i := len(t.state.orders) - 1; i >= 0; i-- {
				if t.state.orders[i].Decision == DecisionBuy {
					buyPrice := t.state.orders[i].Price
					sellPrice := action.Price
					action.Profit = (sellPrice - buyPrice) * action.AssetAmount
					break
				}
			}
		}

		t.state.orders = append(t.state.orders, action)
		indicators := map[string]float64{
			"RSI":         currentCandle.RSI,
			"ATR":         currentCandle.ATR,
			"MACD":        currentCandle.MACD,
			"ADX":         currentCandle.ADX,
			"StochasticK": currentCandle.StochasticK,
			"StochasticD": currentCandle.StochasticD,
			"ShortMA":     currentCandle.ShortMA,
			"LongMA":      currentCandle.LongMA,
			"ShortEMA":    currentCandle.ShortEMA,
			"LongEMA":     currentCandle.LongEMA,
			"UpperBB":     currentCandle.UpperBB,
			"LowerBB":     currentCandle.LowerBB,
		}

		if err := t.metricsCollector.SaveIndicatorData(currentCandle.Timestamp, indicators); err != nil {
			t.log.Error("Failed to save indicator data", zap.Error(err))
		}

		if err := t.metricsCollector.SaveTradeMetric(action, indicators); err != nil {
			t.log.Error("Failed to save trade metric", zap.Error(err))
		}
	}

	return &action, nil
}
