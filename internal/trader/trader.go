package trader

import (
	"bytes"
	"cb_grok/internal/exchange"
	"cb_grok/internal/model"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
	"cb_grok/pkg/models"
	"go.uber.org/zap"
)

type TradeMode string
type TradeDecision string
type TradeDecisionTrigger string

const (
	ModeLiveProd   TradeMode = "live_prod"
	ModeLiveDemo   TradeMode = "live_demo"
	ModeSimulation TradeMode = "simulation"

	DecisionBuy  TradeDecision = "buy"
	DecisionSell TradeDecision = "sell"
	DecisionHold TradeDecision = "hold"

	TriggerStopLoss   TradeDecisionTrigger = "stop_loss"
	TriggerTakeProfit TradeDecisionTrigger = "take_profit"
	TriggerSignal     TradeDecisionTrigger = "signal"
)

var (
	defaultSettings = TraderSettings{
		Commission:           0.001,  // 0.1%
		SlippagePercent:      0.001,  // 0.1%
		Spread:               0.0002, // 0.02%
		StopLossMultiplier:   5,
		TakeProfitMultiplier: 5,
	}
)

type Trader interface {
	Setup(params TraderParams)
	Run(mode TradeMode) error
	BacktestAlgo(appliedOHLCV []models.AppliedOHLCV) (*Action, error)
	GetState() State
}

type State interface {
	GetOrders() []Action
	GetOHLCV() []models.OHLCV
	GetPortfolioValue() float64
	GetPortfolioValues() []PortfolioValue
	CalculateWinRate() float64
	CalculateMaxDrawdown() float64
	CalculateSharpeRatio() float64
	GetInitialCapital() float64

	GenerateCharts() (*bytes.Buffer, error)
}

type trader struct {
	model    *model.Model
	strategy strategy.Strategy
	exch     exchange.Exchange
	state    *state
	settings *TraderSettings

	tg  *telegram.TelegramService
	log *zap.Logger
}

func NewTrader(log *zap.Logger, tg *telegram.TelegramService) Trader {
	return &trader{
		log:      log,
		tg:       tg,
		settings: &defaultSettings,
	}
}

func (t *trader) Setup(params TraderParams) {
	t.model = params.Model

	t.strategy = params.Strategy

	t.exch = params.Exchange

	t.state = t.initState(params.InitialCapital)
	if params.Settings != nil {
		t.settings = params.Settings
	}
}

func (t *trader) GetState() State {
	if t.state == nil {
		return nil
	}
	return t.state
}

func (t *trader) initState(initialCapital float64) *state {
	return &state{
		initialCapital: initialCapital,
		cash:           initialCapital,
	}
}
