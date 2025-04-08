package trader

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/model"
	"cb_grok/internal/strategy"
	"cb_grok/internal/telegram"
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
}

type trader struct {
	model    *model.Model
	strategy strategy.Strategy
	exch     exchange.Exchange
	state    *traderState
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

func (t *trader) initState(initialCapital float64) *traderState {
	return &traderState{
		initialCapital: initialCapital,
	}
}
