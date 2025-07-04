package bybit

import (
	"cb_grok/internal/exchange"
	"context"
	"fmt"
)

func (b *bybit) PlaceSpotMarketOrder(symbol string, orderSide exchange.OrderSide, baseQty float64, takeProfit *float64, stopLoss *float64, precision int64) (string, error) {
	orderSideValue := GetBybitOrderSide(orderSide)
	if orderSideValue == "" {
		return "", fmt.Errorf("unsupported order side: %s", orderSide)
	}

	qty := fmt.Sprintf("%.*f", precision, baseQty)
	fmt.Println("qty:", baseQty, qty, precision)

	req := b.client.NewPlaceOrderService("spot", symbol, orderSideValue, "Market", qty)

	if takeProfit != nil {
		req = req.TakeProfit(fmt.Sprintf("%.2f", *takeProfit))
	}
	if stopLoss != nil {
		req = req.StopLoss(fmt.Sprintf("%.2f", *stopLoss))
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
