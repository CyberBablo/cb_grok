package launcher

import (
	"cb_grok/config"
	"cb_grok/internal/candle"
	"cb_grok/internal/exchange"
	"cb_grok/internal/exchange/bybit"
	"cb_grok/internal/metrics"
	"cb_grok/internal/order"
	stageModel "cb_grok/internal/stage/model"
	"cb_grok/internal/strategy"
	"cb_grok/internal/symbol"
	"cb_grok/internal/telegram"
	"cb_grok/internal/trader"
	"cb_grok/pkg/postgres"
	"fmt"
	"go.uber.org/zap"
)

func Launch(
	traderStage stageModel.StageStatus,
	log *zap.Logger,
	tg *telegram.TelegramService,
	cfg *config.Config,
	orderUC order.Order,
	candleRepo candle.Repository,
	metricsDB postgres.Postgres,
	strategyRepo strategy.Repository,
	traderRepo trader.Repository,
	symbolRepo symbol.Repository,
) error {

	activeTraders, err := traderRepo.GetTraderByStage(int(traderStage))
	if err != nil {
		log.Error("Failed to load active traders", zap.Error(err))
		return fmt.Errorf("error to load model: %w", err)
	}

	for _, activeTrader := range activeTraders {

		activeStrategy, err := strategyRepo.GetStrategy(activeTrader.StrategyID)
		if err != nil {
			log.Error("Failed to load active strategy", zap.Error(err))
		}
		var activeExchange exchange.Exchange

		switch traderStage {
		case stageModel.StageDemo:
			activeExchange, err = bybit.NewBybit(cfg.Bybit.APIKey, cfg.Bybit.APISecret, exchange.TradingModeDemo)
		default:
			activeExchange = exchange.NewMockExchange()
		}
		if err != nil {
			log.Error("Failed to load active exchange", zap.Error(err))
		}
		newTrader := trader.NewTrader(log, tg, orderUC, candleRepo)
		activeSymbol, err := symbolRepo.GetSymbolByID(activeTrader.SymbolID)
		if err != nil {
			log.Error("Failed to load active symbol", zap.Error(err))
		}
		newTrader.Setup(trader.Params{
			Symbol:         *activeSymbol,
			StrategyModel:  activeStrategy,
			Exchange:       activeExchange,
			Strategy:       strategy.NewLinearBiasStrategy(),
			Settings:       nil,
			InitialCapital: activeTrader.InitQty,
			Model:          activeTrader,
		})
		activeMetricCollector := metrics.NewDBMetricsCollector(newTrader.GetState(), metricsDB, activeSymbol.Code, log)
		newTrader.SetMetricsCollector(activeMetricCollector)
		fmt.Println("TRADER RUN", activeTrader.ID)
		if traderStage == stageModel.StageDemo {
			go func() {
				err := newTrader.Run(trader.ModeLiveDemo)
				if err != nil {
					log.Error("failed to run trader", zap.Error(err))
				}
			}()
		} else {
			go func() {
				err := newTrader.RunSimulation(trader.ModeSimulation)
				if err != nil {
					log.Error("failed to run trader", zap.Error(err))
				}
			}()
		}
		fmt.Println("TRADER SET", activeTrader.ID)
	}
	select {}
}
