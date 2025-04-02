package models

type OHLCV struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

type AppliedOHLCV struct {
	OHLCV
	Signal     int // 1 - buy, -1 - sell, 0 - hold
	Position   int // Изменение позиции
	ATR        float64
	RSI        float64
	ShortMA    float64
	LongMA     float64
	ShortEMA   float64
	LongEMA    float64
	Trend      bool
	Volatility bool
}
