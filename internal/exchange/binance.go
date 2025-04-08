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

type binanceImpl struct {
	ex  ccxt.Binance
	log *zap.Logger
}

func NewBinance(isDemo bool, apiKey, apiSecret string, proxyUrl string) (Exchange, error) {
	ex := ccxt.NewBinance(map[string]interface{}{
		"apiKey":    apiKey,
		"secret":    apiSecret,
		"httpProxy": proxyUrl,
	})
	if isDemo {
		ex.SetSandboxMode(true)
	}
	return &binanceImpl{ex: ex, log: logger.GetInstance()}, nil
}

func (e *binanceImpl) CreateOrder(symbol, side string, amount float64, stopLoss, takeProfit float64) error {
	return nil
}

func (e *binanceImpl) FetchOHLCV(symbol string, timeframe string, total int) ([]models.OHLCV, error) {
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

	if len(candles) > total {
		candles = candles[len(candles)-total:]
	}

	sort.Slice(candles, func(i, j int) bool {
		return candles[i].Timestamp < candles[j].Timestamp
	})

	return candles, nil
}

func (e *binanceImpl) GetWSUrl() string {
	return ""
}
