package optimize

import (
	"bytes"
	"cb_grok/config"
	"cb_grok/internal/backtest"
	"cb_grok/internal/exchange"
	"cb_grok/internal/model"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/c-bata/goptuna"
	"github.com/c-bata/goptuna/tpe"
	"github.com/dnlo/struct2csv"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"time"
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
		o.log.Error("Validation days must be greater than 0")
		return fmt.Errorf("validation days must be greater than 0")
	}

	ex, err := exchange.NewBinance(false, o.cfg.Binance.ApuPublic, o.cfg.Binance.ApiSecret, o.cfg.Binance.ProxyUrl)
	if err != nil {
		o.log.Error("optimize: initialize Binance exchange", zap.Error(err))
		return err
	}

	timeframeSec := utils.TimeframeToMilliseconds(params.Timeframe) / 1000
	timePeriodMultiplier := float64(60 * 60 / timeframeSec)
	candlesPerDay := (24 * 60 * 60) / int(timeframeSec)

	candlesTotal := (params.ObjTrainingSetDays + params.ObjValidationSetDays + params.ValidationSetDays) * candlesPerDay

	candles, err := ex.FetchOHLCV(params.Symbol, params.Timeframe, candlesTotal)

	if err != nil {
		o.log.Error("optimize: fetch ohlcv", zap.Error(err))
		return err
	}

	o.log.Info("optimize: ohlcv data", zap.Int("length", len(candles)))

	valCandlesCount := params.ValidationSetDays * candlesPerDay
	objValCandlesCount := params.ObjValidationSetDays * candlesPerDay
	objTrainCandlesCount := params.ObjTrainingSetDays * candlesPerDay

	if (valCandlesCount + objValCandlesCount + objTrainCandlesCount) > len(candles) {
		return fmt.Errorf("summary sets is larger than the available data")
	}

	objTrainCandles := candles[:objTrainCandlesCount]
	objValCandles := candles[objTrainCandlesCount : len(candles)-valCandlesCount]
	valCandles := candles[len(candles)-valCandlesCount:]

	o.log.Info("optimize: datasets prepared",
		zap.Int("obj_train_candles", len(objTrainCandles)),
		zap.Int("obj_val_candles", len(objValCandles)),
		zap.Int("val_candles", len(valCandles)),
	)

	study, err := goptuna.CreateStudy(
		"strategy_1",
		goptuna.StudyOptionDirection(goptuna.StudyDirectionMaximize),
		goptuna.StudyOptionSampler(tpe.NewSampler()),
	)
	if err != nil {
		o.log.Error("optimize: create study", zap.Error(err))
		return err
	}

	eg, ctx := errgroup.WithContext(context.Background())
	study.WithContext(ctx)

	for i := 0; i < params.Workers; i++ {
		eg.Go(func() error {
			return study.Optimize(o.objective(objectiveParams{
				symbol:               params.Symbol,
				trainCandles:         objTrainCandles,
				validationCandles:    objValCandles,
				timePeriodMultiplier: timePeriodMultiplier,
				validationSetDays:    params.ValidationSetDays,
			}), params.Trials/params.Workers)
		})
	}

	if err = eg.Wait(); err != nil {
		o.log.Error("Optimize error %v", zap.Error(err))
		return err
	}

	bestParams, err := study.GetBestParams()
	if err != nil {
		o.log.Error("optimize: get best params", zap.Error(err))
		return err
	}
	combinedSharpeRatio, err := study.GetBestValue()
	if err != nil {
		o.log.Error("optimize: get best value", zap.Error(err))
		return err
	}

	b, err := json.Marshal(bestParams)
	if err != nil {
		o.log.Error("optimize: best params marshal", zap.Error(err))
		return err
	}

	var bestStrategyParams strategy.StrategyParams
	err = json.Unmarshal(b, &bestStrategyParams)
	if err != nil {
		o.log.Error("optimize: best params marshal", zap.Error(err))
		return err
	}

	valBTResult, err := o.bt.Run(valCandles, &model.Model{
		Symbol:         params.Symbol,
		StrategyParams: bestStrategyParams,
	})
	if err != nil {
		o.log.Error("optimize: final validation backtest", zap.Error(err))
		return err
	}

	fmt.Println("ORDER HISTORY")
	for _, order := range valBTResult.Orders {
		o.log.Info(fmt.Sprintf("Order: %v", order))
	}

	filename := model.Save(model.Model{
		Symbol:         params.Symbol,
		StrategyParams: bestStrategyParams,
	})

	orderCount := len(valBTResult.Orders)

	o.log.Info("optimization completed",
		zap.Float64("combined_sharpe_ratio", combinedSharpeRatio),
		zap.Float64("validation_sharpe_ratio", valBTResult.SharpeRatio),
		zap.Float64("validation_max_drawdown", valBTResult.MaxDrawdown),
		zap.Float64("validation_win_rate", valBTResult.WinRate),
		zap.String("filename", filename))

	result := fmt.Sprintf(
		"Символ: %s\nКоличество trials: %d\nКоличество дней на валидации: %d\nTimeframe: %s\nКоличество свечей в сутках: %d\nКоличество сделок: %d\nКомбинированный Sharpe Ratio: %.2f\nВалидационный Sharpe Ratio: %.2f\nИтоговый капитал: %.2f\nМаксимальная просадка: %.2f%%\nWin Rate: %.2f%%\nМодель сохранена в %s",
		params.Symbol, params.Trials, params.ValidationSetDays, params.Timeframe, candlesPerDay, orderCount, combinedSharpeRatio, valBTResult.SharpeRatio, valBTResult.FinalCapital, valBTResult.MaxDrawdown, valBTResult.WinRate, filename)

	if len(valBTResult.Orders) > 0 {
		buff := &bytes.Buffer{}
		w := struct2csv.NewWriter(buff)
		err = w.Write([]string{"timestamp", "side", "trigger", "price", "asset_amount", "asset_currency", "portfolio_value"})
		if err != nil {
			o.log.Error("report: write col names", zap.Error(err))
		}
		for _, v := range valBTResult.Orders {
			var row []string
			row = append(row, time.UnixMilli(v.Timestamp).String())
			row = append(row, string(v.Decision))
			row = append(row, string(v.DecisionTrigger))
			row = append(row, fmt.Sprint(v.Price))
			row = append(row, fmt.Sprint(v.AssetAmount))
			row = append(row, v.AssetCurrency)
			row = append(row, fmt.Sprint(v.PortfolioValue))
			err = w.Write(row)
			if err != nil {
				o.log.Error("report: write structs", zap.Error(err))
			}
		}
		w.Flush()

		err = o.tg.SendFile(buff, "csv", result)
		if err != nil {
			o.log.Error("report: send to telegram", zap.Error(err))
		}
	}

	time.Sleep(1000 * time.Millisecond)
	chartBuff, err := o.generateCharts(valBTResult.TradeState)
	if err != nil {
		o.log.Error("report: generate charts", zap.Error(err))
	}
	err = o.tg.SendFile(chartBuff, "html", "Отчет по симуляции")
	if err != nil {
		o.log.Error("report: send to telegram", zap.Error(err))
	}

	return nil
}
