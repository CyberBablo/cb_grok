package exchange

import (
	"cb_grok/internal/telegram"
	"cb_grok/pkg/models"
	"go.uber.org/zap"
)

type mockImpl struct {
	logger *zap.Logger
	tg     *telegram.TelegramService
}

func NewMockExchange(logger *zap.Logger, tg *telegram.TelegramService) Exchange {
	return &mockImpl{logger: logger, tg: tg}
}

func (e *mockImpl) FetchOHLCV(symbol, timeframe string, limit int) ([]models.OHLCV, error) {
	return nil, nil
}

func (e *mockImpl) CreateOrder(symbol, side string, amount float64, stopLoss, takeProfit float64) error {
	e.logger.Info("Mock order created",
		zap.String("symbol", symbol),
		zap.String("side", side),
		zap.Float64("amount", amount),
		zap.Float64("stop_loss", stopLoss),
		zap.Float64("take_profit", takeProfit))

	e.tg.SendMessage("Mock order created")

	return nil
}

func (e *mockImpl) FetchBalance() (map[string]float64, error) {
	return map[string]float64{}, nil
}

func (e *mockImpl) GetWSUrl() string {
	return "ws://localhost:8080/ws"
}
