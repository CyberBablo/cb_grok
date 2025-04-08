package backtest

type BacktestResult struct {
	SharpeRatio  float64
	Orders       []Order
	FinalCapital float64
	MaxDrawdown  float64
	WinRate      float64
}

type Order struct {
	Action    string  `json:"action"`
	Amount    float64 `json:"amount"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
	Reason    string  `json:"reason,omitempty"`
}
