package optimize

import (
	"cb_grok/config"
	"cb_grok/internal/backtest"
	"cb_grok/internal/exchange"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/trading_model"
	"cb_grok/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/c-bata/goptuna"
	"github.com/c-bata/goptuna/tpe"
	"github.com/ethereum/go-ethereum/log"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Optimize interface {
	Run(params RunOptimizeParams) error
}
type optimize struct {
	log *zap.Logger
	bt  backtest.Backtest
	tg  *telegram.TelegramService
	cfg config.Config
}

func NewOptimize(log *zap.Logger, bt backtest.Backtest, tg *telegram.TelegramService, cfg config.Config) Optimize {
	return &optimize{
		log: log,
		bt:  bt,
		tg:  tg,
		cfg: cfg,
	}
}

func (o *optimize) Run(params RunOptimizeParams) error {
	if params.ValidationSetDays <= 0 {
		log.Error("Validation days must be greater than 0")
		return fmt.Errorf("validation days must be greater than 0")
	}

	ex, err := exchange.NewBinance(false, o.cfg.Binance.ApuPublic, o.cfg.Binance.ApiSecret, o.cfg.Binance.ProxyUrl)
	if err != nil {
		log.Error("optimize: initialize Binance exchange", zap.Error(err))
		return err
	}

	secondsOfDelta := utils.TimeframeToMilliseconds(params.Timeframe) / 1000
	timePeriodMultiplier := float64(60 * 60 / secondsOfDelta)
	candlesPerDay := (24 * 60 * 60) / int(secondsOfDelta)
	candles, err := ex.FetchOHLCV(params.Symbol, params.Timeframe, params.CandlesTotal)

	if err != nil {
		log.Error("optimize: fetch OHLCV", zap.Error(err))
		return err
	}

	log.Info("optimize: OHLCV data", zap.Int("length", len(candles)))

	validationCandlesCount := params.ValidationSetDays * candlesPerDay

	if validationCandlesCount >= len(candles) {
		log.Error("Validation set is larger than the available data")
		return fmt.Errorf("validation set is larger than the available data")
	}

	validationCandles := candles[len(candles)-validationCandlesCount:]
	trainCandles := candles[:len(candles)-validationCandlesCount]

	log.Info("optimize: datasets prepared",
		zap.Int("train_candles", len(trainCandles)),
		zap.Int("validation_candles", len(validationCandles)))

	study, err := goptuna.CreateStudy(
		"strategy_1",
		goptuna.StudyOptionDirection(goptuna.StudyDirectionMaximize),
		goptuna.StudyOptionSampler(tpe.NewSampler()),
	)
	if err != nil {
		log.Error("optimize: create study", zap.Error(err))
		return err
	}

	eg, ctx := errgroup.WithContext(context.Background())
	study.WithContext(ctx)

	for i := 0; i < params.Workers; i++ {
		eg.Go(func() error {
			return study.Optimize(o.objective(objectiveParams{
				trainCandles:         trainCandles,
				validationCandles:    validationCandles,
				timePeriodMultiplier: timePeriodMultiplier,
			}), params.Trials/params.Workers)
		})
	}

	if err = eg.Wait(); err != nil {
		log.Error("Optimize error %v", zap.Error(err))
		return err
	}

	bestParams, err := study.GetBestParams()
	if err != nil {
		log.Error("optimize: get best params", zap.Error(err))
		return err
	}
	combinedSharpeRatio, err := study.GetBestValue()
	if err != nil {
		log.Error("optimize: get best value", zap.Error(err))
		return err
	}

	b, err := json.Marshal(bestParams)
	if err != nil {
		log.Error("optimize: best params marshal", zap.Error(err))
		return err
	}

	var bestStrategyParams strategy.StrategyParams
	err = json.Unmarshal(b, &bestStrategyParams)
	if err != nil {
		log.Error("optimize: best params marshal", zap.Error(err))
		return err
	}

	valBTResult, err := o.bt.Run(validationCandles, bestStrategyParams)
	if err != nil {
		log.Error("optimize: final validation backtest", zap.Error(err))
		return err
	}

	fmt.Println("ORDER HISTORY")
	for _, order := range valBTResult.Orders {
		log.Info(fmt.Sprintf("Order: %v", order))
	}

	filename := trading_model.SaveModel(params.Symbol, bestStrategyParams)

	orderCount := len(valBTResult.Orders)
	result := fmt.Sprintf(
		"Символ: %s\nКоличество trials: %d\nКоличество дней на валидации: %d\nTimeframe: %s\nКоличество свечей в сутках: %d\nКоличество сделок: %d\nКомбинированный Sharpe Ratio: %.2f\nВалидационный Sharpe Ratio: %.2f\nИтоговый капитал: %.2f\nМаксимальная просадка: %.2f%%\nWin Rate: %.2f%%\nМодель сохранена в %s",
		params.Symbol, params.Trials, params.ValidationSetDays, params.Timeframe, candlesPerDay, orderCount, combinedSharpeRatio, valBTResult.SharpeRatio, valBTResult.FinalCapital, valBTResult.MaxDrawdown, valBTResult.WinRate, filename)

	log.Info("optimization completed",
		zap.Float64("combined_sharpe_ratio", combinedSharpeRatio),
		zap.Float64("validation_sharpe_ratio", valBTResult.SharpeRatio),
		zap.Float64("validation_max_drawdown", valBTResult.MaxDrawdown),
		zap.Float64("validation_win_rate", valBTResult.WinRate),
		zap.String("filename", filename))
	o.tg.SendMessage(result)

	return nil
}
