package requestdto

type CancelBet struct {
	BetId      string  `json:"betId"`
	CancelSize float64 `json:"cancelSize"` // ZERO means, cancel everything
}
type CancelBetReqDto struct {
	Token      string `json:"token"`
	OperatorId string `json:"operatorId"`
	ProviderId string `json:"providerId"` // DREAM/BET_FAIR/SPORTS_RADAR
	SportId    string `json:"sportId"`    // CRICKET/SOCCER/TENNIS
	EventId    string `json:"eventId"`
	MarketId   string `json:"marketId"`
	//BetIds     []string `json:"betIds"`
	CancelBets []CancelBet `json:"cancelBets"`
}
