package operatordto

type MarketResultReqDto struct {
	ProviderId   string  `json:"providerId"`   // Dream / BetFair / SportRadar
	SportId      string  `json:"sportId"`      // 1 / 2 / 4
	EventId      string  `json:"eventId"`      // EventId map to gameId in JSON to be inline with GAP
	MarketId     string  `json:"marketId"`     // MarketId map to roundId in JSON to be inline iwth GAP
	MarketType   string  `json:"marketType"`   // MarketType / MATCH_ODDS / BOOKMAKER / FANCY
	Result       string  `json:"result"`       // RunnerId
	ResultName   string  `json:"resultName"`   // RunnerName
	SessionPrice float64 `json:"sessionPrice"` // Session Outcome
}
