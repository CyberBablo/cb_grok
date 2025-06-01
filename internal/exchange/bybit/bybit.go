package bybit

import (
	"cb_grok/internal/exchange"
	"fmt"
	bybitapi "github.com/bybit-exchange/bybit.go.api"
	"go.uber.org/zap"
)

type bybit struct {
	client *bybitapi.Client
	logger *zap.Logger
}

func NewBybit(apiKey, apiSecret string, tradingMode exchange.TradingMode) (exchange.Exchange, error) {
	var clientOptions []bybitapi.ClientOption

	switch tradingMode {
	case exchange.TradingModeDemo:
		clientOptions = append(clientOptions, bybitapi.WithBaseURL(bybitapi.DEMO_ENV), bybitapi.WithDebug(true))
	case exchange.TradingModeTestnet:
		clientOptions = append(clientOptions, bybitapi.WithBaseURL(bybitapi.TESTNET), bybitapi.WithDebug(true))
	case exchange.TradingModeLive:
	default:
		return nil, fmt.Errorf("unsupported trading mode: %s", tradingMode)
	}

	client := bybitapi.NewBybitHttpClient(apiKey, apiSecret, clientOptions...)

	return &bybit{
		client: client,
	}, nil
}
