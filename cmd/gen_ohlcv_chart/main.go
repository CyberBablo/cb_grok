package main

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/utils/logger"
	"cb_grok/pkg/models"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
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

	candles, err := ex.FetchOHLCV("BNB/USDT", "15m", 5760)
	if err != nil {
		log.Error("optimize: fetch ohlcv", zap.Error(err))
		return
	}

	//m, _ := model.Load("model_20250415_162115.json")

	log.Info("fetch ohlcv", zap.Int("length", len(candles)))
	log.Info(fmt.Sprintf("Candles[0]: %+v", candles[0]))
	log.Info(fmt.Sprintf("Candles[%d]: %+v", len(candles)-1, candles[len(candles)-1]))

	//var appliedCandles []models.AppliedOHLCV
	//
	//// 1) full apply
	//strat := strategy.NewLinearBiasStrategy()
	//
	//appliedCandles = strat.ApplyIndicators(candles, m.StrategyParams)

	page := components.NewPage()
	page.AddCharts(
		klineChart(candles),
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

func klineChart(ohlcv []models.OHLCV) *charts.Kline {
	kline := charts.NewKLine()

	x := make([]string, 0)
	y := make([]opts.KlineData, 0)

	// Prepare Kline data and compute price range
	var prices []float64
	for i := 0; i < len(ohlcv); i++ {
		x = append(x, time.UnixMilli(ohlcv[i].Timestamp).Format("2006-01-02 15:04"))
		y = append(y, opts.KlineData{Value: []float64{ohlcv[i].Open, ohlcv[i].Close, ohlcv[i].Low, ohlcv[i].High}})
		prices = append(prices, ohlcv[i].Close)
	}

	// Compute price min/max for scaling indicators
	priceMin, priceMax := prices[0], prices[0]
	for _, p := range prices {
		if p < priceMin {
			priceMin = p
		}
		if p > priceMax {
			priceMax = p
		}
	}
	if priceMax == priceMin {
		priceMax = priceMin + 1
	}

	// Set global options
	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Period history",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: opts.Bool(true),
			Name:  "Price",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:       "slider",
			Start:      0,
			End:        100,
			XAxisIndex: []int{0}, // Zoom only affects X-axis
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:         opts.Bool(true),
			Align:        "right",
			Right:        "0%",
			Orient:       "vertical",
			SelectedMode: "multiple",
			Selected: map[string]bool{
				"kline":       true,
				"ATR":         false,
				"RSI":         false,
				"ShortMA":     false,
				"LongMA":      false,
				"ShortEMA":    false,
				"LongEMA":     false,
				"ADX":         false,
				"MACD":        false,
				"UpperBB":     false,
				"LowerBB":     false,
				"StochasticK": false,
				"StochasticD": false,
			},
		}),
	)

	// Add Kline series
	kline.SetXAxis(x).AddSeries("kline", y)

	return kline
}
