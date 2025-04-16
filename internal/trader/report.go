package trader

import (
	"bytes"
	"fmt"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/samber/lo"
	"io"
	"math"
	"time"
)

func (s *state) GenerateCharts() (*bytes.Buffer, error) {
	page := components.NewPage()
	page.AddCharts(
		s.klineChart(),
		s.lineChart(),
	)

	buff := &bytes.Buffer{}

	err := page.Render(io.MultiWriter(buff))
	if err != nil {
		return nil, err
	}

	return buff, nil
}

// Helper function to map indicator values to price range
func mapToPriceRange(values []float64, priceMin, priceMax float64) []float64 {
	if len(values) == 0 {
		return nil
	}
	// Find min and max of indicator values
	indMin, indMax := values[0], values[0]
	for _, v := range values {
		if !math.IsNaN(v) && !math.IsInf(v, 0) {
			if v < indMin {
				indMin = v
			}
			if v > indMax {
				indMax = v
			}
		}
	}
	// Avoid division by zero
	if indMax == indMin {
		indMax = indMin + 1
	}
	// Map values to price range
	mapped := make([]float64, len(values))
	for i, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			mapped[i] = math.NaN() // Preserve NaN for gaps
		} else {
			mapped[i] = priceMin + ((v-indMin)/(indMax-indMin))*(priceMax-priceMin)
		}
	}
	return mapped
}

func (s *state) klineChart() *charts.Kline {
	kline := charts.NewKLine()

	x := make([]string, 0)
	y := make([]opts.KlineData, 0)
	ohlcv := s.GetOHLCV()

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

	// Add mark points for orders
	var seriesOpts []charts.SeriesOpts
	orders := s.GetOrders()
	for _, v := range orders {
		seriesOpts = append(seriesOpts, charts.WithMarkPointNameCoordItemOpts(opts.MarkPointNameCoordItem{
			Name:       fmt.Sprintf("%s (%s)", v.Decision, v.DecisionTrigger),
			Coordinate: []interface{}{time.UnixMilli(v.Timestamp).Format("2006-01-02 15:04"), v.Price},
			ItemStyle: &opts.ItemStyle{
				Color: lo.If(v.Decision == DecisionBuy, "green").Else("red"),
			},
		}))
	}

	// Add Kline series
	kline.SetXAxis(x).AddSeries("kline", y).SetSeriesOptions(seriesOpts...)

	// Add indicator series if enabled
	appliedOHLCV := s.GetAppliedOHLCV()

	// Prepare indicator data
	atr := make([]opts.LineData, 0)
	rsi := make([]opts.LineData, 0)
	shortMA := make([]opts.LineData, 0)
	longMA := make([]opts.LineData, 0)
	shortEMA := make([]opts.LineData, 0)
	longEMA := make([]opts.LineData, 0)
	adx := make([]opts.LineData, 0)
	macd := make([]opts.LineData, 0)
	upperBB := make([]opts.LineData, 0)
	lowerBB := make([]opts.LineData, 0)
	stochasticK := make([]opts.LineData, 0)
	stochasticD := make([]opts.LineData, 0)

	// Collect raw indicator values for mapping
	var atrValues, rsiValues, shortMAValues, longMAValues, shortEMAValues, longEMAValues []float64
	var adxValues, macdValues, upperBBValues, lowerBBValues, stochasticKValues, stochasticDValues []float64

	for _, v := range appliedOHLCV {
		atr = append(atr, opts.LineData{Value: v.ATR})
		rsi = append(rsi, opts.LineData{Value: v.RSI})
		shortMA = append(shortMA, opts.LineData{Value: v.ShortMA})
		longMA = append(longMA, opts.LineData{Value: v.LongMA})
		shortEMA = append(shortEMA, opts.LineData{Value: v.ShortEMA})
		longEMA = append(longEMA, opts.LineData{Value: v.LongEMA})
		adx = append(adx, opts.LineData{Value: v.ADX})
		macd = append(macd, opts.LineData{Value: v.MACD})
		upperBB = append(upperBB, opts.LineData{Value: v.UpperBB})
		lowerBB = append(lowerBB, opts.LineData{Value: v.LowerBB})
		stochasticK = append(stochasticK, opts.LineData{Value: v.StochasticK})
		stochasticD = append(stochasticD, opts.LineData{Value: v.StochasticD})

		// Collect raw values
		atrValues = append(atrValues, v.ATR)
		rsiValues = append(rsiValues, v.RSI)
		shortMAValues = append(shortMAValues, v.ShortMA)
		longMAValues = append(longMAValues, v.LongMA)
		shortEMAValues = append(shortEMAValues, v.ShortEMA)
		longEMAValues = append(longEMAValues, v.LongEMA)
		adxValues = append(adxValues, v.ADX)
		macdValues = append(macdValues, v.MACD)
		upperBBValues = append(upperBBValues, v.UpperBB)
		lowerBBValues = append(lowerBBValues, v.LowerBB)
		stochasticKValues = append(stochasticKValues, v.StochasticK)
		stochasticDValues = append(stochasticDValues, v.StochasticD)
	}

	// Map indicator values to price range
	atrMapped := mapToPriceRange(atrValues, priceMin, priceMax)
	rsiMapped := mapToPriceRange(rsiValues, priceMin, priceMax)
	shortMAMapped := mapToPriceRange(shortMAValues, priceMin, priceMax)
	longMAMapped := mapToPriceRange(longMAValues, priceMin, priceMax)
	shortEMAMapped := mapToPriceRange(shortEMAValues, priceMin, priceMax)
	longEMAMapped := mapToPriceRange(longEMAValues, priceMin, priceMax)
	adxMapped := mapToPriceRange(adxValues, priceMin, priceMax)
	macdMapped := mapToPriceRange(macdValues, priceMin, priceMax)
	upperBBMapped := mapToPriceRange(upperBBValues, priceMin, priceMax)
	lowerBBMapped := mapToPriceRange(lowerBBValues, priceMin, priceMax)
	stochasticKMapped := mapToPriceRange(stochasticKValues, priceMin, priceMax)
	stochasticDMapped := mapToPriceRange(stochasticDValues, priceMin, priceMax)

	// Update LineData with mapped values
	for i := range appliedOHLCV {
		atr[i] = opts.LineData{Value: atrMapped[i]}
		rsi[i] = opts.LineData{Value: rsiMapped[i]}
		shortMA[i] = opts.LineData{Value: shortMAMapped[i]}
		longMA[i] = opts.LineData{Value: longMAMapped[i]}
		shortEMA[i] = opts.LineData{Value: shortEMAMapped[i]}
		longEMA[i] = opts.LineData{Value: longEMAMapped[i]}
		adx[i] = opts.LineData{Value: adxMapped[i]}
		macd[i] = opts.LineData{Value: macdMapped[i]}
		upperBB[i] = opts.LineData{Value: upperBBMapped[i]}
		lowerBB[i] = opts.LineData{Value: lowerBBMapped[i]}
		stochasticK[i] = opts.LineData{Value: stochasticKMapped[i]}
		stochasticD[i] = opts.LineData{Value: stochasticDMapped[i]}
	}

	// Add indicator series
	indicators := []struct {
		name string
		data []opts.LineData
	}{
		{"ATR", atr},
		//{"RSI", rsi},
		{"ShortMA", shortMA},
		{"LongMA", longMA},
		{"ShortEMA", shortEMA},
		{"LongEMA", longEMA},
		{"ADX", adx},
		{"MACD", macd},
		{"UpperBB", upperBB},
		{"LowerBB", lowerBB},
		{"StochasticK", stochasticK},
		{"StochasticD", stochasticD},
	}

	for _, ind := range indicators {
		// Create a separate Line chart for EMA
		line := charts.NewLine()

		// Overlay the Line chart on the Kline chart
		line.SetXAxis(x).AddSeries(ind.name, ind.data,
			charts.WithLineChartOpts(
				opts.LineChart{
					YAxisIndex: 0, // Use Kline's Y-axis
				},
			),
			//charts.WithLineStyleOpts(
			//	opts.LineStyle{
			//		Color: randomHexColor(),
			//	},
			//),
		)

		kline.Overlap(line)
	}

	return kline
}

func (s *state) lineChart() *charts.Line {
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

	portfolioValues := s.GetPortfolioValues()
	for i := 0; i < len(portfolioValues); i++ {
		x = append(x, time.UnixMilli(portfolioValues[i].Timestamp).Format("2006-01-02 15:04"))
		y = append(y, opts.LineData{Value: portfolioValues[i].Value})
	}

	line.SetXAxis(x).
		AddSeries("Portfolio value", y).
		SetSeriesOptions(
			charts.WithMarkLineNameCoordItemOpts(
				opts.MarkLineNameCoordItem{
					Coordinate0: []interface{}{time.UnixMilli(portfolioValues[0].Timestamp).Format("2006-01-02 15:04"), s.GetInitialCapital()},
					Coordinate1: []interface{}{time.UnixMilli(portfolioValues[len(portfolioValues)-1].Timestamp).Format("2006-01-02 15:04"), s.GetInitialCapital()},
				},
			),
		)

	return line
}
