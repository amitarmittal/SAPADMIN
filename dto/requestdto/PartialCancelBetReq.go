package requestdto

type PartialCancelBetReqDto struct {
	Token      string  `json:"token"`
	OperatorId string  `json:"operatorId"`
	ProviderId string  `json:"providerId"` // DREAM/BET_FAIR/SPORTS_RADAR
	SportId    string  `json:"sportId"`    // CRICKET/SOCCER/TENNIS
	EventId    string  `json:"eventId"`
	MarketId   string  `json:"marketId"`
	BetId      string  `json:"betId"`
	CancelSize float64 `json:"cancelSize"` // 0 means full cancel
}
