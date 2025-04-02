package backtest

import "go.uber.org/fx"

var Module = fx.Module("backtest",
	fx.Provide(NewBacktest),
)
