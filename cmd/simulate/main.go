package main

import (
	"cb_grok/internal/backtest"
	"cb_grok/internal/exchange"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"fmt"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		exchange.Module,
		strategy.Module,
		backtest.Module,
		telegram.Module,
		fx.Invoke(runSimulation),
	).Run()
}

func runSimulation(
	ex exchange.Exchange,
	strat strategy.Strategy,
	bt backtest.Backtest,
	tg *telegram.TelegramService,
) {
	candles, err := ex.Load("BTC/USDT", "1h", 1000)
	if err != nil {
		tg.SendMessage(fmt.Sprintf("Ошибка загрузки данных: %v", err))
		return
	}

	params := strategy.StrategyParams{
		MAShortPeriod:    10,
		MALongPeriod:     20,
		RSIPeriod:        14,
		ATRPeriod:        14,
		BuyRSIThreshold:  30,
		SellRSIThreshold: 70,
		EMAShortPeriod:   50,
		EMALongPeriod:    200,
		UseTrendFilter:   true,
		UseRSIFilter:     true,
		ADXPeriod:        14,
		UseADXFilter:     false,
		ADXThreshold:     25.0,
		ATRThreshold:     0.0,
	}

	sharpeRatio, err := bt.Run(candles, params)
	if err != nil {
		tg.SendMessage(fmt.Sprintf("Ошибка бэктеста: %v", err))
		return
	}

	tg.SendMessage(fmt.Sprintf("Результат симуляции: Sharpe Ratio = %.2f", sharpeRatio))
}
