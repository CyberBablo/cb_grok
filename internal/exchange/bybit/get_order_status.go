package bybit

import (
	order_model "cb_grok/internal/models/order"
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
)

func (b *bybit) GetOrderStatus(orderId string) (order_model.OrderStatus, error) {
	params := map[string]interface{}{"orderId": orderId, "category": "spot"}
	response, err := b.client.NewUtaBybitServiceWithParams(params).GetOpenOrders(context.Background())
	if err != nil {
		b.logger.Error("failed to get order info", zap.String("orderId", orderId), zap.Error(err))
		return 0, err
	}
	fmt.Println(response)
	result, err := ParseResponse(response)
	if err != nil {
		return 0, err
	}

	var orderList OrderList
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal response: %w", err)
	}
	err = json.Unmarshal(resultBytes, &orderList)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(orderList.List) != 1 {
		return 0, fmt.Errorf("expected one order in response, got %d", len(orderList.List))
	}

	ord := orderList.List[0]

	if ord.OrderId != orderId {
		return 0, fmt.Errorf("orderId mismatch: expected %s, got %s", orderId, ord.OrderId)
	}

	status, err := ParseOrderStatus(ord.OrderStatus)
	if err != nil {
		return 0, fmt.Errorf("failed to parse order status: %w", err)
	}

	return status, nil
}
