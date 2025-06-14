package bybit

type WSKlineMessage struct {
	Success bool   `json:"success,omitempty"`
	Type    string `json:"type,omitempty"`
	Topic   string `json:"topic,omitempty"`
	Data    []struct {
		Start     int64  `json:"start"`
		End       int64  `json:"end"`
		Interval  string `json:"interval"`
		Open      string `json:"open"`
		Close     string `json:"close"`
		High      string `json:"high"`
		Low       string `json:"low"`
		Volume    string `json:"volume"`
		Turnover  string `json:"turnover"`
		Confirm   bool   `json:"confirm"`
		Timestamp int64  `json:"timestamp"`
	} `json:"data,omitempty"`
	Ts int64 `json:"ts,omitempty"`
}

type OrderListResult struct {
	List []Order `json:"list,omitempty"`
}

type Order struct {
	OrderId      string `json:"orderId"`
	Symbol       string `json:"symbol"`
	Price        string `json:"price"`
	Qty          string `json:"qty"`
	Side         string `json:"side"`
	OrderStatus  string `json:"orderStatus"`
	CancelType   string `json:"cancelType"`
	RejectReason string `json:"rejectReason"`
	AvgPrice     string `json:"avgPrice"`
	OrderType    string `json:"orderType"`
	TakeProfit   string `json:"takeProfit"`
	StopLoss     string `json:"stopLoss"`
	CreatedTime  string `json:"createdTime"`
	UpdatedTime  string `json:"updatedTime"`
}
