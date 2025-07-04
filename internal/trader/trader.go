package trader

import (
	"bytes"
	"cb_grok/internal/candle"
	"cb_grok/internal/exchange"
	"cb_grok/internal/order"
	"cb_grok/internal/strategy"
	strategyModel "cb_grok/internal/strategy/model"
	symbolModel "cb_grok/internal/symbol/model"
	"cb_grok/internal/telegram"
	traderModel "cb_grok/internal/trader/model"
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
	defaultSettings = Settings{
		Commission:           0.001,  // 0.1%
		SlippagePercent:      0.001,  // 0.1%
		Spread:               0.0002, // 0.02%
		StopLossMultiplier:   5,
		TakeProfitMultiplier: 30,
	}
)

type Trader interface {
	Setup(params Params)
	Run(mode TradeMode) error
	RunSimulation(mode TradeMode) error
	BacktestAlgo(appliedOHLCV []models.AppliedOHLCV) (*Action, error)
	GetState() State
	SetMetricsCollector(collector MetricsCollector)
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
	GetAppliedOHLCV() []models.AppliedOHLCV
	GenerateCharts() (*bytes.Buffer, error)
}

type trader struct {
	model          *traderModel.Trader
	strategyEntity *strategyModel.Strategy
	strategy       strategy.Strategy
	exch           exchange.Exchange
	state          *state
	settings       *Settings
	symbol         symbolModel.Symbol

	orderUC    order.Order
	candleRepo candle.Repository

	tg  *telegram.TelegramService
	log *zap.Logger

	metricsCollector MetricsCollector
}

type MetricsCollector interface {
	SaveTradeMetric(order Action, indicators map[string]float64) error
	SaveIndicatorData(timestamp int64, indicators map[string]float64) error
	Close() error
}

func NewTrader(
	log *zap.Logger,
	tg *telegram.TelegramService,
	orderUC order.Order,
	candleRepo candle.Repository,
) Trader {
	return &trader{
		log:        log,
		tg:         tg,
		settings:   &defaultSettings,
		orderUC:    orderUC,
		candleRepo: candleRepo,
	}
}

func (t *trader) Setup(params Params) {
	t.strategyEntity = params.StrategyModel
	t.strategy = params.Strategy
	t.model = params.Model
	t.exch = params.Exchange
	t.symbol = params.Symbol
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
