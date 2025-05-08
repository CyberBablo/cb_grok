package ml

import "go.uber.org/fx"

var Module = fx.Module("ml",
	fx.Provide(NewMLService),
)
