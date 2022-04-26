package commondto

type ListMarketRespDto []struct {
	MarketID              string  `json:"marketId"`
	Status                string  `json:"status"`
	BetDelay              int     `json:"betDelay"`
	BspReconciled         bool    `json:"bspReconciled"`
	Complete              bool    `json:"complete"`
	Inplay                bool    `json:"inplay"`
	NumberOfWinners       int     `json:"numberOfWinners"`
	NumberOfRunners       int     `json:"numberOfRunners"`
	NumberOfActiveRunners int     `json:"numberOfActiveRunners"`
	TotalMatched          float64 `json:"totalMatched"`
	TotalAvailable        float64 `json:"totalAvailable"`
	CrossMatching         bool    `json:"crossMatching"`
	RunnersVoidable       bool    `json:"runnersVoidable"`
	Version               int64   `json:"version"`
	Runners               []struct {
		SelectionID int     `json:"selectionId"`
		Handicap    float64 `json:"handicap"`
		Status      string  `json:"status"`
	} `json:"runners"`
	MarketDataDelayed bool `json:"marketDataDelayed"`
}
