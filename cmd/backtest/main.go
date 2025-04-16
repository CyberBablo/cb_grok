package main

import (
	"cb_grok/config"
	"cb_grok/internal/backtest"
	"cb_grok/internal/exchange"
	"cb_grok/internal/model"
	"cb_grok/internal/telegram"
	"cb_grok/internal/trader"
	"cb_grok/internal/utils"
	"cb_grok/internal/utils/logger"
	"flag"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"time"
)

const (
	stopLossMultiplier    = 5
	takeProfitsMultiplier = 5
)

func main() {
	fx.New(
		logger.Module,
		config.Module,
		backtest.Module,
		telegram.Module,
		trader.Module,
		fx.Invoke(runBacktest),
	).Run()
}

func runBacktest(backtest backtest.Backtest, cfg config.Config, log *zap.Logger, tg *telegram.TelegramService) error {
	var (
		modelFilename string
		setDays       int
		timeframe     string
	)

	flag.StringVar(&timeframe, "timeframe", "", "Timeframe (f.e 1h)")
	flag.IntVar(&setDays, "set-days", 0, "Number of days for trading set")
	flag.StringVar(&modelFilename, "model", "", "Model filename")
	flag.Parse()

	ex, err := exchange.NewBinance(false, cfg.Binance.ApuPublic, cfg.Binance.ApiSecret, cfg.Binance.ProxyUrl)
	if err != nil {
		log.Error("optimize: initialize Binance exchange", zap.Error(err))
		return err
	}

	mod, err := model.Load(modelFilename)
	if err != nil {
		log.Error("Failed to load model params", zap.Error(err))
		return fmt.Errorf("error to load model: %w", err)
	}

	timeframeSec := utils.TimeframeToMilliseconds(timeframe) / 1000
	candlesPerDay := (24 * 60 * 60) / int(timeframeSec)

	candlesTotal := setDays * candlesPerDay

	candles, err := ex.FetchOHLCV(mod.Symbol, timeframe, candlesTotal)

	result, err := backtest.Run(candles, mod)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf(
		"Результат бектеста:\n\nМодель: %s\nСимвол: %s\nTimeframe: %s\nКол-во свечей: %d\nКол-во дней на валидации: %d\nКоличество сделок: %d\nSharpe Ratio: %.2f\nИтоговый капитал: %.2f\nМаксимальная просадка: %.2f%%\nWin Rate: %.2f%%",
		modelFilename, mod.Symbol, timeframe, len(result.TradeState.GetOHLCV()), setDays, len(result.Orders), result.SharpeRatio, result.FinalCapital, result.MaxDrawdown, result.WinRate)

	time.Sleep(1000 * time.Millisecond)
	chartBuff, err := result.TradeState.GenerateCharts()
	if err != nil {
		log.Error("report: generate charts", zap.Error(err))
	}
	err = tg.SendFile(chartBuff, "html", msg)
	if err != nil {
		log.Error("report: send to telegram", zap.Error(err))
	}

	return nil
}
