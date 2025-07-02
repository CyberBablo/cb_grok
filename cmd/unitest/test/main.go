package test

import (
	"cb_grok/config"
	bybit_uc "cb_grok/internal/exchange/bybit"
	"context"
	"fmt"
	bybit "github.com/bybit-exchange/bybit.go.api"
	"os"
)

func CMD() {
	configPath := os.Getenv("CONFIG_PATH")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}
	bybitApp, err := bybit_uc.NewBybit(cfg.Bybit.APIKey, cfg.Bybit.APISecret, "demo")
	if err != nil {
		panic(err)
	}
	amount, err := bybitApp.GetAvailableSpotWalletBalance("BTC")
	if err != nil {
		panic(err)
	}
	fmt.Println("Amount of demo balance: ", amount)

	buyOrderId := "1981146562319092992"
	status, err := bybitApp.GetOrderStatus(buyOrderId)
	if err != nil {
		panic(err)
	}

	quoteQty, err := bybitApp.GetOrderQuoteQty(buyOrderId)
	if err != nil {
		panic(err)
	}

	fmt.Println("order inf", bybitApp)
	fmt.Println("status", status)
	fmt.Println("quote", quoteQty)

	//response, err := bybitApp.PlaceSpotMarketOrder("BTCUSDT", "buy", 100, nil, nil)
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("Order buy response", response)
	//
	//amount, err = bybitApp.GetAvailableSpotWalletBalance("USDT")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Amount of demo balance: ", amount)

	//response, err := bybitApp.PlaceSpotMarketOrder("BTCUSDT", "sell", 0.001854, nil, nil)
	//
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("Order buy response", response)

}

func testRequest() {
	//ws := bybit.NewBybitPublicWebSocket("wss://stream.bybit.com/v5/public/spot", func(message string) error {
	//	fmt.Println("Received:", message)
	//
	//	return nil
	//})
	//_, _ = ws.Connect().SendSubscription([]string{"kline.15.BTCUSDT"})
	//
	//select {}

	// ------------------------
	configPath := os.Getenv("CONFIG_PATH")

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	var clientOptions []bybit.ClientOption

	clientOptions = append(clientOptions, bybit.WithBaseURL(bybit.DEMO_ENV), bybit.WithDebug(true))

	client := bybit.NewBybitHttpClient(cfg.Bybit.APIKey, cfg.Bybit.APISecret, clientOptions...)

	params := map[string]interface{}{"accountType": "UNIFIED"}

	response, err := client.NewUtaBybitServiceWithParams(params).GetAccountWallet(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println(bybit.PrettyPrint(response))
}
