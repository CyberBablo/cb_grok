package main

import (
	"cb_grok/config"
	"cb_grok/internal/exchange"
	"cb_grok/internal/model"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/trader"
	"cb_grok/internal/utils/logger"
	"flag"
	"fmt"
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
		telegram.Module,
		trader.Module,
		fx.Invoke(runTrade),
	).Run()
}

func runTrade(trade trader.Trader, log *zap.Logger, tg *telegram.TelegramService) error {
	var (
		modelFilename string
	)
	flag.StringVar(&modelFilename, "model", "", "Model filename")
	flag.Parse()

	mod, err := model.Load(modelFilename)
	if err != nil {
		log.Error("Failed to load model params", zap.Error(err))
		return fmt.Errorf("error to load model: %w", err)
	}

	mockExch := exchange.NewMockExchange()

	trade.Setup(trader.TraderParams{
		Model:          mod,
		Exchange:       mockExch,
		Strategy:       strategy.NewLinearBiasStrategy(),
		Settings:       nil, // default
		InitialCapital: 10000,
	})

	err = trade.Run(trader.ModeSimulation)
	if err != nil {
		return err
	}

	return nil
}
