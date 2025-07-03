package model

type RunOptimizeParams struct {
	Symbol       string
	Timeframe    string
	TrainSetDays int
	ValSetDays   int

	Trials  int
	Workers int
}
