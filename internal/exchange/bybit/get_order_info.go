package bybit

import (
	"context"
	"fmt"
)

func (b *bybit) GetOrderInfo(orderId string) (string, error) {
	params := map[string]interface{}{"orderId": orderId, "category": "spot"}
	response, err := b.client.NewUtaBybitServiceWithParams(params).GetOpenOrders(context.Background())
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	result, err := ParseResponse(response)
	if err != nil {
		return "", err
	}

	orderID, ok := result["orderId"].(string)
	if !ok {
		return "", fmt.Errorf("orderId not found in response")
	}

	return orderID, nil
}
