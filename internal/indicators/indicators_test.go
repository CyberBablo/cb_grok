package indicators

import (
	"cb_grok/pkg/models"
	"testing"
)

func TestCalculateADX(t *testing.T) {
	candles := []models.OHLCV{
		{Timestamp: 1, Open: 10, High: 12, Low: 9, Close: 11, Volume: 100},
		{Timestamp: 2, Open: 11, High: 13, Low: 10, Close: 12, Volume: 150},
		{Timestamp: 3, Open: 12, High: 15, Low: 11, Close: 14, Volume: 200},
		{Timestamp: 4, Open: 14, High: 16, Low: 13, Close: 15, Volume: 250},
		{Timestamp: 5, Open: 15, High: 17, Low: 14, Close: 16, Volume: 300},
		{Timestamp: 6, Open: 16, High: 18, Low: 15, Close: 17, Volume: 350},
		{Timestamp: 7, Open: 17, High: 19, Low: 16, Close: 18, Volume: 400},
		{Timestamp: 8, Open: 18, High: 20, Low: 17, Close: 19, Volume: 450},
		{Timestamp: 9, Open: 19, High: 21, Low: 18, Close: 20, Volume: 500},
		{Timestamp: 10, Open: 20, High: 22, Low: 19, Close: 21, Volume: 550},
		{Timestamp: 11, Open: 21, High: 23, Low: 20, Close: 22, Volume: 600},
		{Timestamp: 12, Open: 22, High: 24, Low: 21, Close: 23, Volume: 650},
		{Timestamp: 13, Open: 23, High: 25, Low: 22, Close: 24, Volume: 700},
		{Timestamp: 14, Open: 24, High: 26, Low: 23, Close: 25, Volume: 750},
		{Timestamp: 15, Open: 25, High: 27, Low: 24, Close: 26, Volume: 800},
		{Timestamp: 16, Open: 26, High: 28, Low: 25, Close: 27, Volume: 850},
		{Timestamp: 17, Open: 27, High: 29, Low: 26, Close: 28, Volume: 900},
		{Timestamp: 18, Open: 28, High: 30, Low: 27, Close: 29, Volume: 950},
		{Timestamp: 19, Open: 29, High: 31, Low: 28, Close: 30, Volume: 1000},
		{Timestamp: 20, Open: 30, High: 32, Low: 29, Close: 31, Volume: 1050},
		{Timestamp: 21, Open: 31, High: 33, Low: 30, Close: 32, Volume: 1100},
		{Timestamp: 22, Open: 32, High: 34, Low: 31, Close: 33, Volume: 1150},
		{Timestamp: 23, Open: 33, High: 35, Low: 32, Close: 34, Volume: 1200},
		{Timestamp: 24, Open: 34, High: 36, Low: 33, Close: 35, Volume: 1250},
		{Timestamp: 25, Open: 35, High: 37, Low: 34, Close: 36, Volume: 1300},
		{Timestamp: 26, Open: 36, High: 38, Low: 35, Close: 37, Volume: 1350},
		{Timestamp: 27, Open: 37, High: 39, Low: 36, Close: 38, Volume: 1400},
		{Timestamp: 28, Open: 38, High: 40, Low: 37, Close: 39, Volume: 1450},
		{Timestamp: 29, Open: 39, High: 41, Low: 38, Close: 40, Volume: 1500},
		{Timestamp: 30, Open: 40, High: 42, Low: 39, Close: 41, Volume: 1550},
	}

	adx := CalculateADX(candles, 14)

	if len(adx) != len(candles) {
		t.Errorf("Expected ADX length to be %d, got %d", len(candles), len(adx))
	}

	for i, val := range adx {
		if val < 0 || val > 100 {
			t.Errorf("ADX value at index %d is out of range: %f", i, val)
		}
	}

	for i := 0; i < 2*14-1; i++ {
		if adx[i] != 0 {
			t.Errorf("Expected ADX at index %d to be 0, got %f", i, adx[i])
		}
	}

	nonZeroFound := false
	for i := 2*14 - 1; i < len(adx); i++ {
		if adx[i] > 0 {
			nonZeroFound = true
			break
		}
	}
	if !nonZeroFound {
		t.Errorf("Expected some non-zero ADX values after index %d", 2*14-1)
	}

	adx7 := CalculateADX(candles, 7)
	if len(adx7) != len(candles) {
		t.Errorf("Expected ADX length to be %d, got %d", len(candles), len(adx7))
	}
}
