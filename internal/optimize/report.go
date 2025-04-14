package optimize

import (
	"bytes"
	"cb_grok/internal/trader"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/samber/lo"
	"io"
	"time"
)

func klineChart(tradeState trader.State) *charts.Kline {
	kline := charts.NewKLine()

	x := make([]string, 0)
	y := make([]opts.KlineData, 0)

	ohlcv := tradeState.GetOHLCV()

	for i := 0; i < len(ohlcv); i++ {
		x = append(x, time.UnixMilli(ohlcv[i].Timestamp).Format("2006-01-02 15:04"))
		y = append(y, opts.KlineData{Value: []float64{ohlcv[i].Open, ohlcv[i].Close, ohlcv[i].Low, ohlcv[i].High}})
	}

	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Period history",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			SplitNumber: 20,
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Scale: opts.Bool(true),
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Start: 0,
			End:   100,
		}),
	)

	var seriesOpts []charts.SeriesOpts
	orders := tradeState.GetOrders()
	for _, v := range orders {
		seriesOpts = append(seriesOpts, charts.WithMarkPointNameCoordItemOpts(opts.MarkPointNameCoordItem{
			Name:       fmt.Sprintf("%s (%s)", v.Decision, v.DecisionTrigger),
			Coordinate: []interface{}{time.UnixMilli(v.Timestamp).Format("2006-01-02 15:04"), v.Price},
			ItemStyle: &opts.ItemStyle{
				Color: lo.If(v.Decision == trader.DecisionBuy, "green").Else("red"),
			},
		}))
	}

	//seriesOpts = append(seriesOpts, charts.WithMarkPointStyleOpts(
	//	opts.MarkPointStyle{
	//		Symbol:     []string{"pin"},
	//		SymbolSize: 30,
	//		Label: &opts.Label{
	//			Show:      opts.Bool(true),
	//			Formatter: "{b}: Close {c}", // Shows "News Event: Close 50"
	//			Position:  "top",
	//		},
	//	},
	//))

	kline.SetXAxis(x).AddSeries("kline", y).SetSeriesOptions(seriesOpts...)

	return kline
}

func lineChart(tradeState trader.State) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: "Portfolio value", Subtitle: ""}),
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

	portfolioValues := tradeState.GetPortfolioValues()
	for i := 0; i < len(portfolioValues); i++ {
		x = append(x, time.UnixMilli(portfolioValues[i].Timestamp).Format("2006-01-02 15:04"))
		y = append(y, opts.LineData{Value: portfolioValues[i].Value})
	}

	line.SetXAxis(x).
		AddSeries("Portfolio value", y).
		SetSeriesOptions(
			charts.WithMarkLineNameCoordItemOpts(
				opts.MarkLineNameCoordItem{
					Coordinate0: []interface{}{time.UnixMilli(portfolioValues[0].Timestamp).Format("2006-01-02 15:04"), tradeState.GetInitialCapital()},
					Coordinate1: []interface{}{time.UnixMilli(portfolioValues[len(portfolioValues)-1].Timestamp).Format("2006-01-02 15:04"), tradeState.GetInitialCapital()},
				},
			),
		)

	return line
}

func (o *optimize) generateCharts(tradeState trader.State) (*bytes.Buffer, error) {
	page := components.NewPage()
	page.AddCharts(
		klineChart(tradeState),
		lineChart(tradeState),
	)

	buff := &bytes.Buffer{}

	err := page.Render(io.MultiWriter(buff))
	if err != nil {
		return nil, err
	}

	return buff, nil
}
