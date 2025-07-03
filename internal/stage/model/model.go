package model

type StageStatus int

const (
	StageSimulation StageStatus = 1
	StageDemo       StageStatus = 2
	StageProd       StageStatus = 3
)
