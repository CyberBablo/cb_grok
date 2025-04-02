package strategy

import "go.uber.org/fx"

var Module = fx.Module("strategy",
	fx.Provide(NewStrategy),
)
