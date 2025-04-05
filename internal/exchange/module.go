package exchange

import (
	"cb_grok/config"
	"go.uber.org/fx"
)

var Module = fx.Module("exchange",
	fx.Provide(func(cfg config.Config) (Exchange, error) {
		return NewBinance(cfg.Binance.IsDemo, cfg.Binance.ApuPublic, cfg.Binance.ApiSecret, cfg.Binance.ProxyUrl)
	}),
)
