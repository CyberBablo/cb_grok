package trader

import (
	"cb_grok/pkg/models"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/iamjinlei/go-tachart/tachart"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"os"
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
				t.log.Info("trade algo event",
					zap.Int64("timeframe", action.Timestamp),
					zap.String("decision", string(action.Decision)),
					zap.String("trigger", string(action.DecisionTrigger)),
					zap.Float64("asset_amount", action.AssetAmount),
					zap.String("asset_curr", action.AssetCurrency),
					zap.Float64("portfolio_usdt", action.PortfolioValue),
				)
			}
		}
	}

	if mode == ModeSimulation {
		t.log.Info("simulation results", zap.Int("num_orders", len(t.state.orders)), zap.Float64("tpv", tpv))
		//
		//path, err := generateSimulationReport(*t.state)
		//if err != nil {
		//	t.log.Error("generate report", zap.Error(err))
		//	return nil
		//}
		//
		//t.tg.SendFile(
		//	path,
		//	fmt.Sprintf(
		//		"simulation result:\n\nnum_orders: %d\ntpv: %.2f USDT",
		//		len(t.state.orders),
		//		tpv,
		//	),
		//)
	}
	return nil
}

func generateSimulationReport(state state) (string, error) {
	dir := "lib/simulation_report"
	_ = os.Mkdir(dir, os.ModePerm)

	var chartCandles []tachart.Candle
	var events []tachart.Event

	for _, c := range state.ohlcv {
		chartCandles = append(chartCandles, tachart.Candle{
			Label: time.UnixMilli(c.Timestamp).Format("2006-01-02 15:04"),
			O:     c.Open,
			H:     c.High,
			L:     c.Low,
			C:     c.Close,
			V:     c.Volume,
		})
	}

	for _, e := range state.orders {
		events = append(events, tachart.Event{
			Type:        lo.If(e.Decision == DecisionBuy, tachart.Long).Else(tachart.Short),
			Label:       time.UnixMilli(e.Timestamp).Format("2006-01-02 15:04"),
			Description: string(e.DecisionTrigger),
		})
	}

	cfg := tachart.NewConfig().
		SetChartWidth(1080).
		SetChartHeight(800).
		UseRepoAssets() // serving assets file from current repo, avoid network access

	c := tachart.New(*cfg)

	timestamp := time.Now().Format("20060102_150405")
	path := fmt.Sprintf("%s/simulation_%s.html", dir, timestamp)

	err := c.GenStatic(chartCandles, events, path)
	if err != nil {
		return "", err
	}
	return path, nil
}
