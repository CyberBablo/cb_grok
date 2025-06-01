package bybit

import (
	"cb_grok/internal/exchange"
	"context"
	"fmt"
	bybitapi "github.com/bybit-exchange/bybit.go.api"
	"github.com/govalues/decimal"
)

func (b *bybit) PlaceSpotMarketOrder(ctx context.Context, symbol string, orderSide exchange.OrderSide, amount decimal.Decimal) error {
	orderSideValue := GetBybitOrderSide(orderSide)
	if orderSideValue == "" {
		return fmt.Errorf("unsupported order side: %s", orderSide)
	}

	orderResult, err := b.client.NewPlaceOrderService("spot", symbol, orderSideValue, "Market", amount.String()).Do(ctx)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(bybitapi.PrettyPrint(orderResult))

	return nil
}
