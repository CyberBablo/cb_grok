package optimize

type RunOptimizeParams struct {
	Symbol               string
	Timeframe            string
	ObjValidationSetDays int
	ObjTrainingSetDays   int
	ValidationSetDays    int
	Trials               int
	Workers              int
}
