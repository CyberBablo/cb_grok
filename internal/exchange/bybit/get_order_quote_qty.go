package bybit

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"strconv"
)

func (b *bybit) GetOrderQuoteQty(orderId string) (float64, error) {
	params := map[string]interface{}{"orderId": orderId, "category": "spot"}
	response, err := b.client.NewUtaBybitServiceWithParams(params).GetOrderHistory(context.Background())
	if err != nil {
		b.logger.Error("failed to get order info", zap.String("orderId", orderId), zap.Error(err))
		return 0, err
	}
	fmt.Println("Response order quote", response)
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
	var amount float64
	if ord.Side == "Sell" {
		amount, err = strconv.ParseFloat(ord.CumExecValue, 64)
	} else {
		amount, err = strconv.ParseFloat(ord.CumExecQty, 64)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to parse amount of exec %s: %w", orderId, err)
	}

	return amount, nil
}
