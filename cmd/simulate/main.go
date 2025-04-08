package main

import (
	"cb_grok/internal/backtest"
	"cb_grok/internal/exchange"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/utils"
	"cb_grok/internal/utils/logger"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/schollz/progressbar/v3"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	wsAddr = "localhost:8080"
)

func main() {
	fx.New(
		logger.Module,
		strategy.Module,
		backtest.Module,
		telegram.Module,
		fx.Invoke(runSimulation),
	).Run()
}

func runSimulation(
	log *zap.Logger,
) {
	var (
		symbol      string
		timeframe   string
		tradingDays int
	)
	flag.StringVar(&symbol, "symbol", "", "Symbol (f.e BNB/USDT)")
	flag.StringVar(&timeframe, "timeframe", "", "Timeframe (f.e 1h)")
	flag.IntVar(&tradingDays, "trading-days", 0, "Trading days")
	flag.Parse()

	log.Info("starting ws server", zap.String("symbol", symbol), zap.String("timeframe", timeframe), zap.Int("trading-days", tradingDays))

	go func() {
		err := runServer(log, symbol, timeframe, tradingDays)
		if err != nil {
			log.Error(fmt.Sprintf("run ws server error: %s", err.Error()), zap.String("symbol", symbol), zap.String("timeframe", timeframe), zap.Int("trading-days", tradingDays))
			return
		}
	}()

}

// Server function - Entry point 1
func runServer(log *zap.Logger, symbol string, timeframe string, tradingDays int) error {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ex, err := exchange.NewBinance(false, "", "", "")
	if err != nil {
		return err
	}

	candles, err := ex.FetchOHLCV(symbol, timeframe, calculateLimitFromTimeframeAndDays(timeframe, tradingDays))
	if err != nil {
		return err
	}

	log.Info("simulate: OHLCV data", zap.Int("length", len(candles)))

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Sugar().Errorf("simulation server: upgrade error: %s", err.Error())
			return
		}
		defer conn.Close()

		bar := progressbar.Default(int64(len(candles)), "candles")
		for _, v := range candles {
			d, err := json.Marshal(v)
			if err != nil {
				continue
			}
			err = conn.WriteMessage(websocket.TextMessage, d)
			if err != nil {
				log.Sugar().Errorf("simulation server: write message error: %s", err.Error())
				return
			}
			bar.Add(1)

			time.Sleep(10 * time.Millisecond)
		}

		// Отправляем сообщение о закрытии перед завершением
		closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Simulation completed")
		err = conn.WriteMessage(websocket.CloseMessage, closeMsg)
		if err != nil {
			log.Sugar().Errorf("simulation server: failed to send close message: %s", err.Error())
		}
		log.Info("Simulation completed, connection closed")
	})

	log.Sugar().Infof("Starting WebSocket server on %s", wsAddr)
	err = http.ListenAndServe(wsAddr, nil)
	if err != nil {
		return err
	}

	return nil
}

func calculateLimitFromTimeframeAndDays(tf string, daysNum int) int {
	return int((utils.TimeframeToMilliseconds("1d")*int64(daysNum))/utils.TimeframeToMilliseconds(tf)) + 1
}
