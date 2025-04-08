package trader

import "go.uber.org/fx"

var Module = fx.Module("trader",
	fx.Provide(NewTrader),
)
