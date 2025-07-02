package trader

import (
	"bytes"
	"sync"

	"go.uber.org/zap"

	"cb_grok/internal/candle"
	"cb_grok/internal/exchange"
	candle_model "cb_grok/internal/models/candle"
	strategy_model "cb_grok/internal/models/strategy"
	"cb_grok/internal/order"
	"cb_grok/internal/strategy"
	"cb_grok/pkg/telegram"
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
		TakeProfitMultiplier: 30,
	}
)

type Trader interface {
	Setup(params TraderParams)
	Run(mode TradeMode, timeframe string) error
	RunSimulation(mode TradeMode) error
	BacktestAlgo(appliedOHLCV []candle_model.AppliedOHLCV) (*Action, error)
	GetState() State
	SetMetricsCollector(collector MetricsCollector)
}

type State interface {
	GetOrders() []Action
	GetOHLCV() []candle_model.OHLCV
	GetPortfolioValue() float64
	GetPortfolioValues() []PortfolioValue
	CalculateWinRate() float64
	CalculateMaxDrawdown() float64
	CalculateSharpeRatio() float64
	GetInitialCapital() float64
	GetAppliedOHLCV() []candle_model.AppliedOHLCV

	GenerateCharts() (*bytes.Buffer, error)
}

type trader struct {
	model    *strategy_model.StrategyFileModel
	strategy strategy.Strategy
	exch     exchange.Exchange
	state    *state
	settings *TraderSettings

	orderUC    order.Usecase
	candleRepo candle.Repository

	tg  *telegram.TelegramService
	log *zap.Logger

	mu sync.RWMutex

	metricsCollector MetricsCollector
}

type MetricsCollector interface {
	SaveTradeMetric(order Action, indicators map[string]float64) error
	SaveIndicatorData(timestamp int64, indicators map[string]float64) error
	Close() error
}

func NewTrader(log *zap.Logger, tg *telegram.TelegramService, orderUC order.Usecase, candleRepo candle.Repository) Trader {
	return &trader{
		log:        log,
		tg:         tg,
		settings:   &defaultSettings,
		orderUC:    orderUC,
		candleRepo: candleRepo,
	}
}

func (t *trader) Setup(params TraderParams) {
	t.model = params.Model

	t.strategy = params.Strategy

	t.exch = params.Exchange
	t.orderUC.Init(params.Exchange)

	t.state = t.initState(params.InitialCapital)
	if params.Settings != nil {
		t.settings = params.Settings
	}
}

func (t *trader) SetMetricsCollector(collector MetricsCollector) {
	t.metricsCollector = collector
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
	}
}
