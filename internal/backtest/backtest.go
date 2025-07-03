package backtest

import (
	"cb_grok/internal/candle"
	"cb_grok/internal/exchange"
	"cb_grok/internal/model"
	"cb_grok/internal/order"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/internal/trader"
	model2 "cb_grok/internal/trader/model"
	"cb_grok/pkg/models"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Backtest interface {
	Run(candles []models.OHLCV, mod *model.Model) (*BacktestResult, error)
}

type backtest struct {
	InitialCapital       float64
	Commission           float64
	SlippagePercent      float64
	Spread               float64
	StopLossMultiplier   float64
	TakeProfitMultiplier float64

	tg  *telegram.TelegramService
	log *zap.Logger

	orderUC    order.Order
	candleRepo candle.Repository
}

func NewBacktest(log *zap.Logger, tg *telegram.TelegramService, orderUC order.Order, candleRepo candle.Repository) Backtest {
	return &backtest{
		InitialCapital:       10000.0,
		Commission:           0.001,  // 0.1%
		SlippagePercent:      0.001,  // 0.1%
		Spread:               0.0002, // 0.02%
		StopLossMultiplier:   5,
		TakeProfitMultiplier: 30,

		tg:         tg,
		log:        log,
		orderUC:    orderUC,
		candleRepo: candleRepo,
	}
}

func (b *backtest) Run(ohlcv []models.OHLCV, mod *model.Model) (*BacktestResult, error) {
	str := strategy.NewLinearBiasStrategy()

	trade := trader.NewTrader(b.log, b.tg, b.orderUC, b.candleRepo)
	trade.Setup(model2.TraderParams{
		Model:    mod,
		Exchange: exchange.NewMockExchange(),
		Strategy: str,
		Settings: &model2.TraderSettings{
			Commission:           b.Commission,
			SlippagePercent:      b.SlippagePercent,
			Spread:               b.Spread,
			StopLossMultiplier:   b.StopLossMultiplier,
			TakeProfitMultiplier: b.TakeProfitMultiplier,
		},
		InitialCapital: b.InitialCapital,
	})

	appliedCandles := str.ApplyIndicators(ohlcv, mod.StrategyParams)
	if appliedCandles == nil {
		return nil, fmt.Errorf("no candles after strategy apply")
	}

	for _, c := range appliedCandles {
		if c.ATR == 0 {
			return nil, fmt.Errorf("ATR is required for backtest")
		}
	}

	var valCandles []models.AppliedOHLCV

	for i := 0; i < len(appliedCandles); i++ {
		valCandles = append(valCandles, appliedCandles[i])

		_, _ = trade.BacktestAlgo(valCandles)
	}

	tradeState := trade.GetState()

	if len(tradeState.GetOrders()) > 1 {
		return &BacktestResult{
			SharpeRatio:  tradeState.CalculateSharpeRatio(),
			Orders:       tradeState.GetOrders(),
			FinalCapital: tradeState.GetPortfolioValue(),
			MaxDrawdown:  tradeState.CalculateMaxDrawdown(),
			WinRate:      tradeState.CalculateWinRate(),
			TradeState:   tradeState,
		}, nil
	}

	return &BacktestResult{
		Orders:       tradeState.GetOrders(),
		FinalCapital: tradeState.GetPortfolioValue(),
		TradeState:   tradeState,
	}, nil
}

var Module = fx.Module("backtest",
	fx.Provide(NewBacktest),
)
