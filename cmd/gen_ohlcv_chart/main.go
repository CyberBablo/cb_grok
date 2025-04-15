package main

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/model"
	"cb_grok/internal/strategy"
	"cb_grok/internal/utils/logger"
	"cb_grok/pkg/models"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"io"
	"os"
	"time"
)

func main() {
	log := logger.NewLogger()

	ex, err := exchange.NewBinance(false, "", "", "")
	if err != nil {
		log.Error("optimize: initialize Bybit exchange", zap.Error(err))
		return
	}

	candles, err := ex.FetchOHLCV("BNB/USDT", "15m", 3000)
	if err != nil {
		log.Error("optimize: fetch ohlcv", zap.Error(err))
		return
	}

	m, _ := model.Load("model_20250415_162115.json")

	log.Info("fetch ohlcv", zap.Int("length", len(candles)))
	log.Info(fmt.Sprintf("Candles[0]: %+v", candles[0]))
	log.Info(fmt.Sprintf("Candles[%d]: %+v", len(candles)-1, candles[len(candles)-1]))

	var appliedCandles1 []models.AppliedOHLCV
	var appliedCandles2 []models.AppliedOHLCV

	// 1) full apply
	strat := strategy.NewLinearBiasStrategy()

	appliedCandles1 = strat.ApplyIndicators(candles, m.StrategyParams)

	// 2) iterative apply
	var ohlcv []models.OHLCV
	for i := range candles {
		ohlcv = append(ohlcv, candles[i])

		appliedCandles2 = strat.ApplyIndicators(ohlcv, m.StrategyParams)
	}

	page := components.NewPage()
	page.AddCharts(
		lineChart("ATR", appliedCandles1),
		lineChart("ATR", appliedCandles2),

		lineChart("RSI", appliedCandles1),
		lineChart("RSI", appliedCandles2),

		lineChart("ShortMA", appliedCandles1),
		lineChart("ShortMA", appliedCandles2),

		lineChart("LongEMA", appliedCandles1),
		lineChart("LongEMA", appliedCandles2),

		lineChart("Trend", appliedCandles1),
		lineChart("Trend", appliedCandles2),

		lineChart("Volatility", appliedCandles1),
		lineChart("Volatility", appliedCandles2),

		lineChart("ADX", appliedCandles1),
		lineChart("ADX", appliedCandles2),

		lineChart("MACD", appliedCandles1),
		lineChart("MACD", appliedCandles2),

		lineChart("UpperBB", appliedCandles1),
		lineChart("UpperBB", appliedCandles2),

		lineChart("LowerBB", appliedCandles1),
		lineChart("LowerBB", appliedCandles2),

		lineChart("StochasticK", appliedCandles1),
		lineChart("StochasticK", appliedCandles2),

		lineChart("StochasticD", appliedCandles1),
		lineChart("StochasticD", appliedCandles2),
	)

	f, err := os.Create("kline.html")
	if err != nil {
		panic(err)

	}

	err = page.Render(io.MultiWriter(f))
	if err != nil {
		panic(err)
	}
}

func lineChart(name string, data []models.AppliedOHLCV) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: name, Subtitle: ""}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: opts.Bool(true),
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start: 0,
			End:   100,
		}),
	)

	x := make([]string, 0)
	y := make([]opts.LineData, 0)

	for i := 0; i < len(data); i++ {
		x = append(x, time.UnixMilli(data[i].Timestamp).Format("2006-01-02 15:04"))

		value := 0.0
		switch name {
		case "ATR":
			value = data[i].ATR
		case "RSI":
			value = data[i].RSI
		case "ShortMA":
			value = data[i].ShortMA
		case "LongEMA":
			value = data[i].LongEMA
		case "Trend":
			value = lo.If(data[i].Trend, 1.0).Else(0.0)
		case "Volatility":
			value = lo.If(data[i].Volatility, 1.0).Else(0.0)
		case "ADX":
			value = data[i].ADX
		case "MACD":
			value = data[i].MACD
		case "UpperBB":
			value = data[i].UpperBB
		case "LowerBB":
			value = data[i].LowerBB
		case "StochasticK":
			value = data[i].StochasticK
		case "StochasticD":
			value = data[i].StochasticD

		}
		y = append(y, opts.LineData{Value: value})
	}

	line.SetXAxis(x).
		AddSeries(name, y)

	return line
}
