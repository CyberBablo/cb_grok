package bybit

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"strconv"
)

func (b *bybit) GetAvailableSpotWalletBalance(coin string) (float64, error) {
	params := map[string]interface{}{"accountType": "UNIFIED"}
	response, err := b.client.NewUtaBybitServiceWithParams(params).GetAccountWallet(context.Background())
	if err != nil {
		b.logger.Error("failed to get wallet balance", zap.String("coin", coin), zap.Error(err))
		return 0, err
	}

	result, err := ParseResponse(response)
	if err != nil {
		return 0, err
	}

	var walletList WalletBalanceList
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal response: %w", err)
	}
	err = json.Unmarshal(resultBytes, &walletList)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(walletList.List) != 1 {
		return 0, fmt.Errorf("expected one wallet in response, got %d", len(walletList.List))
	}

	for _, c := range walletList.List[0].Coin {
		if c.Coin == coin {
			bal, err := strconv.ParseFloat(c.WalletBalance, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse balance for coin %s: %w", coin, err)
			}
			return bal, nil
		}
	}

	return 0, nil
}
