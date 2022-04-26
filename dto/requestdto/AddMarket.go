package requestdto

type AddMarket struct {
	OperatorId    string `json:"operatorId"`
	ProviderId    string `json:"providerId"`    // Dream / BetFair / SportRadar
	SportId       string `json:"sportId"`       // 1 / 2 / 3
	CompetitionId string `json:"competitionId"` // IPL / BBL
	EventId       string `json:"eventId"`       // Unique Event Id
	MarketId      string `json:"marketId"`      // Unique Market Id
	MarketType    string `json:"marketType"`    // MATCH_ODDS / BOOKMAKER / FANCY
}
