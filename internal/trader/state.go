package trader

import (
	"cb_grok/pkg/models"
	"math"
)

func (s *state) GetOrders() []Action {
	return s.orders
}

func (s *state) GetOHLCV() []models.OHLCV {
	return s.ohlcv
}

func (s *state) GetInitialCapital() float64 {
	return s.initialCapital
}

func (s *state) GetPortfolioValues() []PortfolioValue {
	return s.portfolioValues
}

func (s *state) GetPortfolioValue() float64 {
	if len(s.portfolioValues) <= 0 {
		return s.initialCapital
	}
	return s.portfolioValues[len(s.portfolioValues)-1].Value
}

func (s *state) CalculateWinRate() float64 {
	if len(s.orders) == 0 {
		return 0.0
	}

	var wins int
	for i := 1; i < len(s.orders); i += 2 { // Проходим парами buy-sell
		buy := s.orders[i-1]
		sell := s.orders[i]

		if sell.Decision == DecisionSell && buy.Decision == DecisionBuy {
			buyPrice := buy.Price
			sellPrice := sell.Price
			if sellPrice > buyPrice {
				wins++
			}
		}
	}

	totalTrades := len(s.orders) / 2
	if totalTrades == 0 {
		return 0.0
	}

	return float64(wins) / float64(totalTrades) * 100.0
}

func (s *state) CalculateMaxDrawdown() float64 {
	if len(s.portfolioValues) == 0 {
		return 0.0
	}

	peak := s.initialCapital
	maxDrawdown := 0.0

	// Вычисляем максимальную просадку
	for _, value := range s.portfolioValues {
		if value.Value > peak {
			peak = value.Value
		}
		drawdown := (peak - value.Value) / peak * 100.0
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

func (s *state) CalculateSharpeRatio() float64 {
	if len(s.orders) < 2 {
		return 0.0
	}

	// Собираем доходности
	var returns []float64
	for i := 1; i < len(s.portfolioValues); i++ {
		prevValue := s.portfolioValues[i-1]
		currValue := s.portfolioValues[i]
		if prevValue.Value != 0 {
			returns = append(returns, (currValue.Value-prevValue.Value)/prevValue.Value)
		}
	}

	if len(returns) == 0 {
		return 0.0
	}

	// Вычисляем среднюю доходность
	sum := 0.0
	for _, r := range returns {
		sum += r
	}
	mean := sum / float64(len(returns))

	// Вычисляем стандартное отклонение
	sumSquaredDiff := 0.0
	for _, r := range returns {
		sumSquaredDiff += math.Pow(r-mean, 2)
	}
	stdDev := math.Sqrt(sumSquaredDiff / float64(len(returns)))

	if stdDev == 0 {
		return 0.0
	}

	// Годовой Шарп (дневные данные, ~252 торговых дня в году)
	return mean / stdDev * math.Sqrt(252)
}
