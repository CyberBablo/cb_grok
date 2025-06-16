package bybit

import (
	"cb_grok/internal/exchange"
	"context"
	"fmt"
	"strconv"
)

func (b *bybit) PlaceSpotMarketOrder(symbol string, orderSide exchange.OrderSide, baseQty float64, takeProfit *float64, stopLoss *float64) (string, error) {
	orderSideValue := GetBybitOrderSide(orderSide)
	if orderSideValue == "" {
		return "", fmt.Errorf("unsupported order side: %s", orderSide)
	}

	qty := strconv.FormatFloat(baseQty, 'f', -1, 64)

	req := b.client.NewPlaceOrderService("spot", symbol, orderSideValue, "Market", qty)

	if takeProfit != nil {
		req.TakeProfit(fmt.Sprintf("%.2f", *takeProfit))
	}
	if stopLoss != nil {
		req.StopLoss(fmt.Sprintf("%.2f", *stopLoss))
	}

	orderResult, err := req.Do(context.Background())
	if err != nil {
		return "", err
	}

	result, err := ParseResponse(orderResult)
	if err != nil {
		return "", err
	}

	orderID, ok := result["orderId"].(string)
	if !ok {
		return "", fmt.Errorf("orderId not found in response")
	}

	return orderID, nil
}
