package bybit

import (
	"cb_grok/internal/exchange"
	"context"
	"fmt"
	bybitapi "github.com/bybit-exchange/bybit.go.api"
	"strconv"
)

func (b *bybit) PlaceSpotMarketOrder(symbol string, orderSide exchange.OrderSide, orderAmount float64) error {
	orderSideValue := GetBybitOrderSide(orderSide)
	if orderSideValue == "" {
		return fmt.Errorf("unsupported order side: %s", orderSide)
	}

	orderAmountValue := strconv.FormatFloat(orderAmount, 'f', -1, 64)

	orderResult, err := b.client.NewPlaceOrderService("spot", symbol, orderSideValue, "Market", orderAmountValue).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(bybitapi.PrettyPrint(orderResult))

	return nil
}
