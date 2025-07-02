// Package ohlc provides functionality to generate OHLC charts and visualizations
// from cryptocurrency trading data.
package ohlc

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"cb_grok/config"
	"cb_grok/internal/exchange"
	"cb_grok/internal/exchange/bybit"
	candle_model "cb_grok/internal/models/candle"
	"cb_grok/pkg/logger"
)

// OHLCParams holds the command line parameters for OHLC operations.
type OHLCParams struct {
	Symbol    string
	Timeframe string
	Count     int
	Output    string
	Title     string
}

// CMD runs the OHLC chart generation command with FX dependency injection.
func CMD() {
	// Parse command line flags
	params := parseOHLCFlags()

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create FX application
	app := fx.New(
		// Provide configuration
		fx.Provide(func() *config.Config { return cfg }),
		fx.Provide(func() OHLCParams { return params }),

		// Provide logger
		fx.Provide(func(cfg *config.Config) (*zap.Logger, error) {
			return logger.NewZapLogger(logger.ZapConfig{
				Level:        cfg.Logger.Level,
				Development:  cfg.Logger.Development,
				Encoding:     cfg.Logger.Encoding,
				OutputPaths:  cfg.Logger.OutputPaths,
				FileLog:      cfg.Logger.FileLog,
				FilePath:     cfg.Logger.FilePath,
				FileMaxSize:  cfg.Logger.FileMaxSize,
				FileCompress: cfg.Logger.FileCompress,
			})
		}),

		// Provide exchange
		fx.Provide(func() (exchange.Exchange, error) {
			return bybit.NewBybit("", "", "demo")
		}),

		// Lifecycle management
		fx.Invoke(runOHLCOperation),

		// FX logger
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)

	app.Run()
}

// createKlineChart creates a candlestick chart from OHLCV data with professional styling.
func createKlineChart(ohlcv []candle_model.OHLCV, title string) *charts.Kline {
	kline := charts.NewKLine()

	x := make([]string, 0)
	y := make([]opts.KlineData, 0)

	// Prepare Kline data and compute price range
	var prices []float64
	for i := 0; i < len(ohlcv); i++ {
		x = append(x, time.UnixMilli(ohlcv[i].Timestamp).Format("2006-01-02 15:04"))

		y = append(y, opts.KlineData{Value: []float64{ohlcv[i].Close, ohlcv[i].Open, ohlcv[i].Low, ohlcv[i].High}})
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

	// Set global options with professional styling
	kline.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: title,
			Subtitle: fmt.Sprintf("Period: %s - %s | Total Candles: %d",
				time.UnixMilli(ohlcv[0].Timestamp).Format("Jan 02, 2006"),
				time.UnixMilli(ohlcv[len(ohlcv)-1].Timestamp).Format("Jan 02, 2006"),
				len(ohlcv)),
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

// parseOHLCFlags parses command line flags and returns OHLCParams.
func parseOHLCFlags() OHLCParams {
	var (
		symbol    = flag.String("symbol", "BNBUSDT", "Trading symbol (e.g., BTCUSDT, ETHUSDT)")
		timeframe = flag.String("timeframe", "15m", "Timeframe for candles (1m, 5m, 15m, 1h, etc.)")
		count     = flag.Int("count", 1200, "Number of candles to fetch")
		output    = flag.String("output", "kline.html", "Output HTML file name")
		title     = flag.String("title", "", "Chart title (auto-generated if empty)")
	)
	flag.Parse()

	return OHLCParams{
		Symbol:    *symbol,
		Timeframe: *timeframe,
		Count:     *count,
		Output:    *output,
		Title:     *title,
	}
}

// runOHLCOperation is the main FX lifecycle function that executes OHLC operations.
func runOHLCOperation(
	lifecycle fx.Lifecycle,
	log *zap.Logger,
	ex exchange.Exchange,
	params OHLCParams,
	shutdowner fx.Shutdowner,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Info("Starting OHLC chart generation with FX",
				zap.String("symbol", params.Symbol),
				zap.String("timeframe", params.Timeframe),
				zap.Int("count", params.Count),
				zap.String("output", params.Output))

			go func() {
				defer shutdowner.Shutdown()

				// Fetch candle data
				log.Info("Fetching candle data...")
				candles, err := ex.FetchSpotOHLCV(params.Symbol, exchange.Timeframe(params.Timeframe), params.Count)
				if err != nil {
					log.Error("Failed to fetch OHLCV data", zap.Error(err))
					return
				}

				log.Info("Fetched candle data successfully",
					zap.Int("candles_count", len(candles)),
					zap.String("first_candle", time.UnixMilli(candles[0].Timestamp).Format("2006-01-02 15:04")),
					zap.String("last_candle", time.UnixMilli(candles[len(candles)-1].Timestamp).Format("2006-01-02 15:04")))

				// Generate chart title if not provided
				chartTitle := params.Title
				if chartTitle == "" {
					chartTitle = fmt.Sprintf("%s %s Price Chart", params.Symbol, params.Timeframe)
				}

				// Create and render chart
				log.Info("Generating chart...")
				page := components.NewPage()
				page.AddCharts(createKlineChart(candles, chartTitle))

				// Save to file
				if err := saveChart(page, params.Output, log); err != nil {
					log.Error("Failed to save chart", zap.Error(err))
					return
				}

				log.Info("Chart generated successfully", zap.String("output_file", params.Output))
				fmt.Printf("✅ OHLC chart saved to: %s\n", params.Output)
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("OHLC chart generation completed")
			return nil
		},
	})
}

// saveChart saves the chart page to an HTML file with proper error handling.
func saveChart(page *components.Page, filename string, log *zap.Logger) error {
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file '%s': %w", filename, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Warn("Failed to close file", zap.String("filename", filename), zap.Error(closeErr))
		}
	}()

	if err := page.Render(io.MultiWriter(f)); err != nil {
		return fmt.Errorf("failed to render chart to file '%s': %w", filename, err)
	}

	return nil
}
