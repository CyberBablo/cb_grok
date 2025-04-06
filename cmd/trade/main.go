package main

import (
	"cb_grok/config"
	"cb_grok/internal/exchange"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/utils/logger"
	"cb_grok/pkg/models"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/iamjinlei/go-tachart/tachart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const (
	stopLossMultiplier    = 5
	takeProfitsMultiplier = 5
)

func main() {
	fx.New(
		logger.Module,
		config.Module,
		exchange.Module,
		telegram.Module,
		fx.Invoke(runLiveTrading),
	).Run()
}

func runLiveTrading(log *zap.Logger, tg *telegram.TelegramService) error {
	var (
		modelFile      string
		mode           string
		initialCapital float64
	)
	flag.StringVar(&modelFile, "model-file", "", "Model filename")
	flag.StringVar(&mode, "mode", "", "Trading mode (simulation, demo, prod)")
	flag.Float64Var(&initialCapital, "initial-capital", 10000, "Initial capital")
	flag.Parse()

	var strategyParams strategy.StrategyParams

	modelParams, err := loadModelParams(modelFile, &strategyParams)
	if err != nil {
		log.Error("Failed to load model params", zap.Error(err))
		return err
	}

	symbol := modelParams["symbol"].(string)

	var ex exchange.Exchange
	if mode == "simulation" {
		ex = exchange.NewMockExchange(log, tg)
	} else {
		return errors.New("unsupported live-trading mode")
	}

	// Определение WebSocket URL
	wsUrl, err := determineWSURL(mode)
	if err != nil {
		return err
	}
	log.Info("Connecting to WebSocket", zap.String("url", wsUrl), zap.String("mode", mode))
	tg.SendMessage(fmt.Sprintf("Подключение к %s в режиме %s", wsUrl, mode))

	// Подключение к WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		log.Error("WebSocket connection failed", zap.Error(err))
		tg.SendMessage(fmt.Sprintf("Ошибка WebSocket: %v", err))
		return err
	}
	defer conn.Close()

	// Инициализация состояния
	var dataBuffer []models.OHLCV
	cash := initialCapital
	assets := 0.0
	positionOpen := false
	entryPrice := 0.0
	stopLoss := 0.0
	takeProfit := 0.0
	requiredCandles := max(strategyParams.EMALongPeriod, strategyParams.MALongPeriod, strategyParams.MACDLongPeriod, strategyParams.ATRPeriod, strategyParams.RSIPeriod)

	var events []tachart.Event
	// Основной цикл обработки сообщений
	for {
		_, message, err := conn.ReadMessage()
		if err != nil && websocket.IsCloseError(err, websocket.CloseNormalClosure) {
			log.Info("WebSocket closed normally", zap.Error(err))
			tg.SendMessage("Симуляция завершена")
			break
		}
		if err != nil {
			log.Error("WebSocket read error", zap.Error(err))
			tg.SendMessage(fmt.Sprintf("Ошибка WebSocket: %v", err))
			break
		}

		var candle models.OHLCV
		err = json.Unmarshal(message, &candle)
		if err != nil {
			log.Error("cannot parse ohlcv message", zap.Error(err))
		}

		// Накопление данных
		dataBuffer = append(dataBuffer, candle)

		if len(dataBuffer) < requiredCandles {
			continue
		}

		// Применение стратегии
		strat := strategy.NewMovingAverageStrategy()
		processedCandles := strat.Apply(dataBuffer, strategyParams)
		if processedCandles == nil {
			log.Warn("Strategy application failed")
			continue
		}

		latestCandle := processedCandles[len(processedCandles)-1]
		latestSignal := latestCandle.Signal
		currentPrice := latestCandle.Close
		atr := latestCandle.ATR

		decision := "Hold"
		var event tachart.EventType
		transactionAmount := 0.0

		// Логика торговли
		if positionOpen {
			if currentPrice <= stopLoss {
				decision = "Sell (stop-loss)"
				event = tachart.Short
				transactionAmount = assets

				err = ex.CreateOrder(symbol, "sell", assets, stopLoss, takeProfit)
				if err != nil {
					log.Error("create order failed", zap.Error(err))
				}
				cash += assets * currentPrice

				assets = 0.0
				positionOpen = false
			} else if currentPrice >= takeProfit {
				decision = "Sell (take-profit)"
				event = tachart.Short
				transactionAmount = assets

				err = ex.CreateOrder(symbol, "sell", assets, stopLoss, takeProfit)
				if err != nil {
					log.Error("create order failed", zap.Error(err))
				}
				cash += assets * currentPrice

				assets = 0.0
				positionOpen = false
			} else if latestSignal == -1 {
				decision = "Sell (signal)"
				event = tachart.Short
				transactionAmount = assets

				err = ex.CreateOrder(symbol, "sell", assets, stopLoss, takeProfit)
				if err != nil {
					log.Error("create order failed", zap.Error(err))
				}
				cash += assets * currentPrice

				assets = 0.0
				positionOpen = false
			}
		} else if latestSignal == 1 && cash > 0 {
			decision = "Buy"
			event = tachart.Long
			transactionAmount = cash / currentPrice

			err = ex.CreateOrder(symbol, "buy", transactionAmount, stopLoss, takeProfit)
			assets = transactionAmount
			cash = 0.0

			positionOpen = true
			entryPrice = currentPrice
			stopLoss = entryPrice - atr*stopLossMultiplier
			takeProfit = entryPrice + atr*takeProfitsMultiplier
		}

		if err != nil {
			log.Error("Order creation failed", zap.Error(err))
			tg.SendMessage(fmt.Sprintf("Ошибка создания ордера: %v", err))
			continue
		}

		// Формирование отчета
		portfolioValue := cash + assets*currentPrice
		actionDetail := ""
		if decision != "Hold" {
			actionDetail = fmt.Sprintf("%s: %.2f %s", decision, transactionAmount, strings.Split(symbol, "/")[0])
		}
		portfolioDetail := fmt.Sprintf("(%.2f USDT + %.2f %s)", cash, assets, strings.Split(symbol, "/")[0])
		msg := fmt.Sprintf("[%s] Symbol: %s, Decision: %s, Price: %.2f, %s, Portfolio: %.2f USDT %s",
			time.UnixMilli(candle.Timestamp).Format(time.RFC3339), symbol, decision, currentPrice, actionDetail, portfolioValue, portfolioDetail)
		_ = msg

		if decision != "Hold" {
			events = append(events, tachart.Event{
				Type:        event,
				Label:       time.UnixMilli(candle.Timestamp).Format("2006-01-02 15:04"),
				Description: fmt.Sprintf("Decision: %s", decision),
			})
		}

	}

	if mode == "simulation" {
		log.Info("generating report", zap.Int("candles_length", len(dataBuffer)), zap.Int("event_length", len(events)))
		err := generateSimulationReport(dataBuffer, events)
		if err != nil {
			log.Error("generate report", zap.Error(err))
		}
	}
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

func loadModelParams(filename string, strategyParams *strategy.StrategyParams) (map[string]interface{}, error) {
	path := filepath.Join("lib/best_models", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var params map[string]interface{}

	err = json.Unmarshal(data, strategyParams)
	if err != nil {
		return nil, err
	}

	return params, json.Unmarshal(data, &params)
}

func determineWSURL(mode string) (string, error) {
	switch mode {
	case "simulation":
		return "ws://localhost:8080/ws", nil
	}
	return "", errors.New("unsupported trading mode")
}
