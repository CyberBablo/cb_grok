package telegram

import (
	"go.uber.org/fx"
)

var Module = fx.Module("telegram",
	fx.Provide(NewTelegramService),
)
