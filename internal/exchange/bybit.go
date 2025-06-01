package exchange

import (
	"cb_grok/internal/utils"
	"cb_grok/internal/utils/logger"
	"cb_grok/pkg/models"
	"fmt"
	ccxt "github.com/ccxt/ccxt/go/v4"
	"go.uber.org/zap"
	"sort"
)

type bybitImpl struct {
	ex  ccxt.Bybit
	log *zap.Logger
}

func NewBybit(isDemo bool, apiKey, apiSecret string) (Exchange, error) {
	ex := ccxt.NewBybit(map[string]interface{}{
		"apiKey": apiKey,
		"secret": apiSecret,
	})
	if isDemo {
		ex.SetSandboxMode(true)
	}
	return &bybitImpl{ex: ex, log: logger.GetInstance()}, nil
}

func (e *bybitImpl) CreateOrder(symbol, side string, amount float64, stopLoss, takeProfit float64) error {
	return nil
}

func (e *bybitImpl) FetchOHLCV(symbol string, timeframe string, total int) ([]models.OHLCV, error) {
	limit := 1000

	var candles []models.OHLCV

	since := int64(0)

	for {
		options := []ccxt.FetchOHLCVOptions{
			ccxt.WithFetchOHLCVTimeframe(timeframe),
			ccxt.WithFetchOHLCVLimit(int64(limit)),
		}
		if since != 0 {
			options = append(options, ccxt.WithFetchOHLCVSince(since))
		}

		ohlcv, err := e.ex.FetchOHLCV(
			symbol, options...,
		)
		if err != nil {
			return nil, err
		}

		for _, c := range ohlcv {
			candles = append(candles, models.OHLCV{
				Timestamp: c.Timestamp,
				Open:      c.Open,
				High:      c.High,
				Low:       c.Low,
				Close:     c.Close,
				Volume:    c.Volume,
			})
		}

		if len(ohlcv) < limit {
			break
		}

		if len(candles) >= total {
			break
		}

		since = ohlcv[0].Timestamp - (int64(limit) * utils.TimeframeToMilliseconds(timeframe))

		e.log.Info(fmt.Sprintf("fetched ohlcv: %d/%d", len(candles), total))
	}

	sort.Slice(candles, func(i, j int) bool {
		return candles[i].Timestamp < candles[j].Timestamp
	})

	if len(candles) > 1 {
		uniqueCandles := []models.OHLCV{candles[0]}
		for i := 1; i < len(candles); i++ {
			if candles[i].Timestamp != candles[i-1].Timestamp {
				uniqueCandles = append(uniqueCandles, candles[i])
			}
		}
		candles = uniqueCandles
	}

	if len(candles) > total {
		candles = candles[len(candles)-total:]
	}
	return candles, nil
}

func (e *bybitImpl) GetWSUrl() string {
	return ""
}
