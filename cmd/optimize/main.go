package main

import (
	"cb_grok/config"
	"cb_grok/internal/backtest"
	"cb_grok/internal/optimize"
	"cb_grok/internal/telegram"
	"cb_grok/internal/utils/logger"
	"flag"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		logger.Module,
		backtest.Module,
		telegram.Module,
		config.Module,
		optimize.Module,
		fx.Invoke(runOptimization),
	)

	app.Run()
}

func runOptimization(opt optimize.Optimize) error {
	var (
		validationSetDays int
		symbol            string
		timeframe         string
		trials            int
		candlesTotal      int
		workers           int
	)

	flag.StringVar(&symbol, "symbol", "", "Symbol (f.e BNB/USDT)")
	flag.StringVar(&timeframe, "timeframe", "", "Timeframe (f.e 1h)")
	flag.IntVar(&trials, "trials", 100, "Number of trials")
	flag.IntVar(&validationSetDays, "val-set-days", 0, "Number of days for validation set")
	flag.IntVar(&candlesTotal, "candles-total", 1000, "Count of candles in the dataset")
	flag.IntVar(&workers, "workers", 2, "Number of parallel workers")

	flag.Parse()

	return opt.Run(optimize.RunOptimizeParams{
		Symbol:            symbol,
		Timeframe:         timeframe,
		ValidationSetDays: validationSetDays,
		Trials:            trials,
		CandlesTotal:      candlesTotal,
		Workers:           workers,
	})
}

//func registerOptimize(
//	lifeCycle fx.Lifecycle,
//	opt optimize.Optimize,
//	shutdowner fx.Shutdowner,
//) {
//	lifeCycle.Append(fx.Hook{
//		OnStart: func(_ context.Context) error {
//			if err := runOptimization(opt); err != nil {
//				log.Fatalf("run optimize error: %v", err)
//			}
//			time.Sleep(1 * time.Second) // TODO: implement graceful-shutdown support
//			if err := shutdowner.Shutdown(); err != nil {
//				log.Printf("shutdown error: %v", err)
//			}
//			return nil
//		},
//		OnStop: func(_ context.Context) error {
//			return nil
//		},
//	})
//}
