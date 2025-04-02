package exchange

import "go.uber.org/fx"

var Module = fx.Module("exchange",
	fx.Provide(NewBybit),
)
