package place_order

import (
	"cb_grok/config"
	"context"
	"fmt"
	bybit "github.com/bybit-exchange/bybit.go.api"
	"os"
)

func CMD() {
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
