package betfair

type Runner struct {
	RunnerId     string `json:"runnerId"`
	RunnerStatus string `json:"runnerStatus"`
}

type MarketResult struct {
	MarketId     string   `json:"marketId"`
	MarketStatus string   `json:"marketStatus"`
	Runners      []Runner `json:"runners"`
}

type GetMarketResultResp struct {
	Status       string       `json:"status"`
	Message      string       `json:"message"`
	MarketResult MarketResult `json:"marketResult"`
}
