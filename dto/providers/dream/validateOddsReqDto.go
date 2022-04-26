package dream

type ValidateOddsReqDto struct {
	SportId        string  `json:"sportId"`
	EventId        string  `json:"eventId"`
	MarketType     string  `json:"marketType"`
	MarketId       string  `json:"marketId"`
	RunnerId       string  `json:"runnerId"`
	BetType        string  `json:"oddType"`
	OddValue       float64 `json:"oddValue"`
	SessionOutcome float64 `json:"sessionRuns"`
}
