package optimize

type RunOptimizeParams struct {
	Symbol            string
	Timeframe         string
	ValidationSetDays int
	Trials            int
	CandlesTotal      int
	Workers           int
}
