package model

import "time"

type StrategyParams struct {
	MAShortPeriod       int     `json:"ma_short_period"`
	MALongPeriod        int     `json:"ma_long_period"`
	RSIPeriod           int     `json:"rsi_period"`
	ATRPeriod           int     `json:"atr_period"`
	BuyRSIThreshold     float64 `json:"buy_rsi_threshold"`
	SellRSIThreshold    float64 `json:"sell_rsi_threshold"`
	EMAShortPeriod      int     `json:"ema_short_period"`
	EMALongPeriod       int     `json:"ema_long_period"`
	ATRThreshold        float64 `json:"atr_threshold"`
	MACDShortPeriod     int     `json:"macd_short_period"`
	MACDLongPeriod      int     `json:"macd_long_period"`
	MACDSignalPeriod    int     `json:"macd_signal_period"`
	EMAWeight           float64 `json:"ema_weight"`
	TrendWeight         float64 `json:"trend_weight"`
	RSIWeight           float64 `json:"rsi_weight"`
	MACDWeight          float64 `json:"macd_weight"`
	BuySignalThreshold  float64 `json:"buy_signal_threshold"`
	SellSignalThreshold float64 `json:"sell_signal_threshold"`
	BollingerPeriod     int     `json:"bollinger_period"`
	BollingerStdDev     float64 `json:"bollinger_std_dev"`
	BBWeight            float64 `json:"bb_weight"`
	StochasticKPeriod   int     `json:"stochastic_k_period"` // Период для %K
	StochasticDPeriod   int     `json:"stochastic_d_period"` // Период для %D
	StochasticWeight    float64 `json:"stochastic_weight"`   // Вес для сигнала
}

type Signals struct {
	EMASignal        int
	RSISignal        int
	MACDSignal       int
	TrendSignal      int
	BBSignal         int
	StochasticSignal int // Сигнал от Stochastic
}

type Strategy struct {
	ID        int            `db:"id"`
	CreatedAt time.Time      `db:"created_at"`
	SymbolID  int            `db:"symbol_id"`
	Params    StrategyParams `db:"params"`
	TimeFrame string         `db:"timeframe"`
}
