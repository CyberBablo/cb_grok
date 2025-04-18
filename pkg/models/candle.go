package models

type OHLCV struct {
	Timestamp int64   `json:"timestamp"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Close     float64 `json:"close"`
	Volume    float64 `json:"volume"`
}

type AppliedOHLCV struct {
	OHLCV
	ATR         float64
	RSI         float64
	ShortMA     float64
	LongMA      float64
	ShortEMA    float64
	LongEMA     float64
	Trend       bool
	Volatility  bool
	ADX         float64
	MACD        float64
	MACDSignal  float64
	UpperBB     float64
	LowerBB     float64
	StochasticK float64
	StochasticD float64
	OBV         float64 // Новое поле
	VWAP        float64 // Новое поле
	CCI         float64 // Новое поле
	Signal      int
}

//type AppliedOHLCV struct {
//	OHLCV
//	ATR         float64
//	RSI         float64
//	ShortMA     float64
//	LongMA      float64
//	ShortEMA    float64
//	LongEMA     float64
//	Trend       bool
//	Volatility  bool
//	ADX         float64
//	MACD        float64
//	MACDSignal  float64
//	Signal      int
//	UpperBB     float64
//	LowerBB     float64
//	StochasticK float64 // Значение %K
//	StochasticD float64 // Значение %D
//}
