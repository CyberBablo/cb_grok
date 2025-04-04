package main

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/utils/logger"
	"fmt"
	"go.uber.org/zap"
	"time"

	"github.com/iamjinlei/go-tachart/tachart"
)

func main() {
	log := logger.NewLogger()

	ex, err := exchange.NewBinance(false, "", "")
	if err != nil {
		log.Error("optimize: initialize Bybit exchange", zap.Error(err))
		return
	}

	candles, err := ex.FetchOHLCV("BNB/USDT", "1h", 3000)
	if err != nil {
		log.Error("optimize: fetch OHLCV", zap.Error(err))
		return
	}

	log.Info("fetch OHLCV", zap.Int("length", len(candles)))
	log.Info(fmt.Sprintf("Candles[0]: %+v", candles[0]))
	log.Info(fmt.Sprintf("Candles[%d]: %+v", len(candles)-1, candles[len(candles)-1]))

	var chartCandles []tachart.Candle

	for _, c := range candles[:100] {
		chartCandles = append(chartCandles, tachart.Candle{
			Label: time.UnixMilli(c.Timestamp).Format("2006-01-02 15:04"),
			O:     c.Open,
			H:     c.High,
			L:     c.Low,
			C:     c.Close,
			V:     c.Volume,
		})
	}

	events := []tachart.Event{
		{
			Type:        tachart.Short,
			Label:       chartCandles[40].Label,
			Description: "This is a demo event description. Randomly pick this candle to go short on " + chartCandles[40].Label,
		},
	}

	cfg := tachart.NewConfig().
		SetChartWidth(1080).
		SetChartHeight(800).
		AddOverlay(
			tachart.NewSMA(5),
			tachart.NewSMA(20),
		).
		AddIndicator(
			tachart.NewMACD(12, 26, 9),
		).
		UseRepoAssets() // serving assets file from current repo, avoid network access

	c := tachart.New(*cfg)
	c.GenStatic(chartCandles, events, "kline.html")
}
