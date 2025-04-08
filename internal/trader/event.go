package trader

import "fmt"

func (e *Event) String() string {
	return fmt.Sprintf("Timestamp: %d, Decision: %s, Trigger: %s, Assets: %.2f %s, Current portfolio: %.2f USDT",
		e.Timestamp, e.Decision, e.DecisionTrigger, e.AssetAmount, e.AssetCurrency, e.PortfolioValue)
}
