package operatordto

type MarketRollbackReqDto struct {
	ProviderId   string `json:"providerId"`   // Dream / BetFair / SportRadar
	SportId      string `json:"sportId"`      // EventId map to gameId in JSON to be inline with GAP
	EventId      string `json:"eventId"`      // EventId map to gameId in JSON to be inline with GAP
	MarketId     string `json:"marketId"`     // MarketId map to roundId in JSON to be inline iwth GAP
	MarketName   string `json:"marketName"`   // MarketType / MATCH_ODDS / BOOKMAKER / FANCY
	MarketType   string `json:"marketType"`   // MarketType / MATCH_ODDS / BOOKMAKER / FANCY
	Reason       string `json:"reason"`       // MarketType / MATCH_ODDS / BOOKMAKER / FANCY
	RollbackType string `json:"rollbackType"` // RunnerId
}
