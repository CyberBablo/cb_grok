package trader

import traderModel "cb_grok/internal/trader/model"

type Repository interface {
	GetTraderByStage(stageID int) ([]*traderModel.Trader, error)
}
