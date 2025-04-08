package trader

import (
	"cb_grok/pkg/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/iamjinlei/go-tachart/tachart"
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

	for {
		_, message, err := conn.ReadMessage()
		if err != nil && websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			t.log.Info("WebSocket closed normally", zap.Error(err))
			//t.tg.SendMessage("Симуляция завершена")
			break
		}
		if err != nil {
			t.log.Error("WebSocket read error", zap.Error(err))
			t.tg.SendMessage(fmt.Sprintf("Ошибка WebSocket: %v", err))
			break
		}

		var candle models.OHLCV
		err = json.Unmarshal(message, &candle)
		if err != nil {
			t.log.Error("cannot parse ohlcv message", zap.Error(err))
			continue
		}

		action, err := t.algo(candle)
		if err != nil {
			t.log.Error("trade algo error", zap.Error(err), zap.String("message", string(message)))
			t.tg.SendMessage(fmt.Sprintf("trade algo error: %s\n\nMessage: %s", err.Error(), string(message)))
		}

		if action != nil && action.Decision == DecisionBuy {
			b, _ := json.Marshal(action)

			fmt.Printf(">>>>> %s\n", string(message))

			fmt.Printf("%s\n\n", string(b))
		}

	}

	t.log.Info("simulation results", zap.Int("num_events", len(t.state.events)))

	//if mode == "simulation" {
	//	log.Info("generating report", zap.Int("candles_length", len(dataBuffer)), zap.Int("event_length", len(events)))
	//	err := generateSimulationReport(dataBuffer, events)
	//	if err != nil {
	//		log.Error("generate report", zap.Error(err))
	//	}
	//}
	return nil
}

func generateSimulationReport(candles []models.OHLCV, events []tachart.Event) error {
	var chartCandles []tachart.Candle

	for _, c := range candles {
		chartCandles = append(chartCandles, tachart.Candle{
			Label: time.UnixMilli(c.Timestamp).Format("2006-01-02 15:04"),
			O:     c.Open,
			H:     c.High,
			L:     c.Low,
			C:     c.Close,
			V:     c.Volume,
		})
	}

	cfg := tachart.NewConfig().
		SetChartWidth(1080).
		SetChartHeight(800).
		UseRepoAssets() // serving assets file from current repo, avoid network access

	c := tachart.New(*cfg)
	timestamp := time.Now().Format("20060102_150405")
	err := c.GenStatic(chartCandles, events, fmt.Sprintf("simulation_%s.html", timestamp))
	if err != nil {
		return err
	}
	return nil
}
