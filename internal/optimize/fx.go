package optimize

import "go.uber.org/fx"

var Module = fx.Module("optimize",
	fx.Provide(NewOptimize),
)
