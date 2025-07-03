package bybit

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/order/model"
	"fmt"
	bybitapi "github.com/bybit-exchange/bybit.go.api"
)

func GetBybitOrderSide(orderSide exchange.OrderSide) string {
	switch orderSide {
	case exchange.OrderSideBuy:
		return "Buy"
	case exchange.OrderSideSell:
		return "Sell"
	default:
		return ""
	}
}

func GetBybitTimeframe(timeframe exchange.Timeframe) string {
	switch timeframe {
	case exchange.Timeframe1m:
		return "1"
	case exchange.Timeframe5m:
		return "5"
	case exchange.Timeframe15m:
		return "15"
	case exchange.Timeframe30m:
		return "30"
	case exchange.Timeframe1h:
		return "60"
	case exchange.Timeframe4h:
		return "240"
	case exchange.Timeframe1d:
		return "D"
	case exchange.Timeframe1w:
		return "W"
	case exchange.Timeframe1M:
		return "M"
	default:
		return ""
	}
}

func ParseResponse(response *bybitapi.ServerResponse) (map[string]interface{}, error) {
	if response == nil {
		return nil, fmt.Errorf("response is empty")
	}
	if response.RetMsg != "OK" || response.RetCode != 0 {
		return nil, fmt.Errorf("request failed: %s, code: %d", response.RetMsg, response.RetCode)
	}

	return response.Result.(map[string]interface{}), nil
}

func ParseOrderStatus(status string) (order_model.OrderStatus, error) {
	switch status {
	case "New", "PartiallyFilled":
		return order_model.OrderStatusPlaced, nil
	case "Filled":
		return order_model.OrderStatusFilled, nil
	case "Cancelled", "Rejected", "Deactivated":
		return order_model.OrderStatusCanceled, nil
	default:
		return 0, fmt.Errorf("unknown order status: %s", status)
	}
}
