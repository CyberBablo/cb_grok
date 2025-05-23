package trader

import (
	"bytes"
	"cb_grok/pkg/models"
	"encoding/json"
	"fmt"
	"github.com/dnlo/struct2csv"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"time"
)

func (t *trader) Run(mode TradeMode) error {
	if t.model == nil || t.state == nil || t.settings == nil {
		return fmt.Errorf("required fields are empty. Setup the trader first")
	}
	if mode != ModeSimulation {
		return fmt.Errorf("unsupported trade mode")
	}

	wsUrl := t.exch.GetWSUrl()

	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		t.log.Error("WebSocket connection failed", zap.Error(err))
		t.tg.SendMessage(fmt.Sprintf("Ошибка WebSocket: %v", err))
		return err
	}
	defer conn.Close()

	t.log.Info("connected to websocket", zap.String("url", wsUrl))

	b, _ := json.MarshalIndent(t.model, "", "    ")
	fmt.Printf("Use model:\n%+v\n", string(b))

	// --------------------------------------------------------------

	tpv := 0.0

	for {
		_, message, err := conn.ReadMessage()
		if err != nil && websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			t.log.Info("websocket closed normally", zap.Error(err))
			break
		}
		if err != nil {
			t.log.Error("websocket read error", zap.Error(err))
			t.tg.SendMessage(fmt.Sprintf("websocket error: %v", err))
			break
		}

		var candle models.OHLCV
		err = json.Unmarshal(message, &candle)
		if err != nil {
			t.log.Error("cannot parse ohlcv message", zap.Error(err))
			continue
		}

		action, err := t.processAlgo(candle)
		if err != nil {
			t.log.Error("trade algo error", zap.Error(err), zap.String("message", string(message)))
			t.tg.SendMessage(fmt.Sprintf("trade algo error: %s\n\nMessage: %s", err.Error(), string(message)))
		}

		if action != nil {
			tpv = action.PortfolioValue

			if action.Decision != DecisionHold {
				//t.log.Info("trade algo event",
				//	zap.Int64("timeframe", action.Timestamp),
				//	zap.String("decision", string(action.Decision)),
				//	zap.String("trigger", string(action.DecisionTrigger)),
				//	zap.Float64("asset_amount", action.AssetAmount),
				//	zap.String("asset_curr", action.AssetCurrency),
				//	zap.Float64("portfolio_usdt", action.PortfolioValue),
				//)
			}
		}
	}

	if mode == ModeSimulation {
		t.log.Info("simulation results", zap.Int("num_orders", len(t.state.orders)), zap.Float64("tpv", tpv))

		result := fmt.Sprintf(
			"Результат симуляции\n\nСимвол: %s\nКол-во свечей: %d\nКоличество сделок: %d\nSharpe Ratio: %.2f\nИтоговый капитал: %.2f\nМаксимальная просадка: %.2f%%\nWin Rate: %.2f",
			t.model.Symbol, len(t.state.GetOHLCV()), len(t.state.GetOrders()), t.state.CalculateSharpeRatio(), t.state.GetPortfolioValue(), t.state.CalculateMaxDrawdown(), t.state.CalculateWinRate())

		buff := &bytes.Buffer{}
		w := struct2csv.NewWriter(buff)
		err = w.Write([]string{"timestamp", "side", "trigger", "price", "asset_amount", "asset_currency", "portfolio_value"})
		if err != nil {
			t.log.Error("report: write col names", zap.Error(err))
		}
		for _, v := range t.state.GetOrders() {
			var row []string
			row = append(row, time.UnixMilli(v.Timestamp).String())
			row = append(row, string(v.Decision))
			row = append(row, string(v.DecisionTrigger))
			row = append(row, fmt.Sprint(v.Price))
			row = append(row, fmt.Sprint(v.AssetAmount))
			row = append(row, v.AssetCurrency)
			row = append(row, fmt.Sprint(v.PortfolioValue))
			err = w.Write(row)
			if err != nil {
				t.log.Error("report: write structs", zap.Error(err))
			}
		}
		w.Flush()

		err = t.tg.SendFile(buff, "csv", result)
		if err != nil {
			t.log.Error("report: send to telegram", zap.Error(err))
		}

		time.Sleep(1000 * time.Millisecond)
		chartBuff, err := t.state.GenerateCharts()
		if err != nil {
			t.log.Error("report: generate charts", zap.Error(err))
		}
		err = t.tg.SendFile(chartBuff, "html", "Отчет по симуляции")
		if err != nil {
			t.log.Error("report: send to telegram", zap.Error(err))
		}
	}
	return nil
}
