package bybit

import (
	"cb_grok/internal/exchange"
	"cb_grok/internal/utils"
	"cb_grok/pkg/models"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	bybitapi "github.com/bybit-exchange/bybit.go.api"
	"go.uber.org/zap"
	"sort"
	"strconv"
)

func (b *bybit) FetchSpotOHLCV(symbol string, timeframe exchange.Timeframe, total int) ([]models.OHLCV, error) {
	timeframeValue := GetBybitTimeframe(timeframe)
	if timeframeValue == "" {
		return nil, fmt.Errorf("unsupported timeframe: %s", timeframe)
	}

	limit := 1000
	if total < limit {
		limit = total
	}

	var candles []models.OHLCV

	since := int64(0)

	for {
		params := map[string]interface{}{"category": "spot", "symbol": symbol, "interval": timeframeValue, "limit": limit}
		if since != 0 {
			params["start"] = since
		}
		response, err := b.client.NewUtaBybitServiceWithParams(params).GetMarketKline(context.Background())
		if err != nil {
			fmt.Println(err)
			return nil, errors.New("failed to fetch ohlcv: " + err.Error())
		}
		_, err = ParseResponse(response)
		if err != nil {
			return nil, err
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}
		result, _, err := bybitapi.GetMarketKlineResponse(nil, responseBytes, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		ohlcv := result.List

		for _, r := range ohlcv {
			ts, err := strconv.ParseInt(r.StartTime, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}
			o, err := strconv.ParseFloat(r.OpenPrice, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse decimal in ohlcv: %w", err)
			}
			h, err := strconv.ParseFloat(r.HighPrice, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse decimal in ohlcv: %w", err)
			}
			l, err := strconv.ParseFloat(r.LowPrice, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse decimal in ohlcv: %w", err)
			}
			c, err := strconv.ParseFloat(r.ClosePrice, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse decimal in ohlcv: %w", err)
			}
			v, err := strconv.ParseFloat(r.Volume, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse decimal in ohlcv: %w", err)
			}
			candles = append(candles, models.OHLCV{
				Timestamp: ts,
				Open:      o,
				High:      h,
				Low:       l,
				Close:     c,
				Volume:    v,
			})
		}

		if len(result.List) < limit {
			break
		}

		if len(candles) >= total {
			break
		}

		startTimestamp, err := strconv.ParseInt(ohlcv[len(ohlcv)-1].StartTime, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse last candle timestamp: %w", err)
		}

		since = startTimestamp - (int64(limit) * utils.TimeframeToMilliseconds(string(timeframe)))

		zap.L().Info(fmt.Sprintf("fetched ohlcv: %d/%d", len(candles), total))

	}

	// sort ASC
	sort.Slice(candles, func(i, j int) bool {
		return candles[i].Timestamp < candles[j].Timestamp
	})

	// deduplication
	if len(candles) > 1 {
		uniqueCandles := []models.OHLCV{candles[0]}
		for i := 1; i < len(candles); i++ {
			if candles[i].Timestamp != candles[i-1].Timestamp {
				uniqueCandles = append(uniqueCandles, candles[i])
			}
		}
		candles = uniqueCandles
	}

	// limit count to total
	if len(candles) > total {
		candles = candles[len(candles)-total:]
	}

	return candles, nil
}
