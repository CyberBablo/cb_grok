package exchange

type OrderSide string

var (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

type TradingMode string

var (
	TradingModeDemo    TradingMode = "demo"
	TradingModeLive    TradingMode = "live"
	TradingModeTestnet TradingMode = "testnet"
)

type Timeframe string

var (
	Timeframe1m  Timeframe = "1m"
	Timeframe5m  Timeframe = "5m"
	Timeframe15m Timeframe = "15m"
	Timeframe30m Timeframe = "30m"
	Timeframe1h  Timeframe = "1h"
	Timeframe4h  Timeframe = "4h"
	Timeframe1d  Timeframe = "1d"
	Timeframe1w  Timeframe = "1w"
	Timeframe1M  Timeframe = "1M"
)
