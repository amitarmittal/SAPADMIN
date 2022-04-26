package operatordto

type OddsData struct {
	OddsKey   string  `json:"oddsKey"`
	OddsAt    string  `json:"oddsAt"`
	OddsValue float64 `json:"oddsValue"`
}
type BetUpdate struct {
	Token           string     `json:"token"`
	BetId           string     `json:"betId"`
	BetPlacedTime   string     `json:"betPlacedTime"`
	BetStatus       string     `json:"betStatus"`
	BetType         string     `json:"betType"`
	SportId         string     `json:"sportId"`
	EventId         string     `json:"eventId"`
	MarketId        string     `json:"marketId"`
	MarketName      string     `json:"marketName"`
	MarketType      string     `json:"marketType"`
	RunnerId        string     `json:"runnerId"`
	RunnerName      string     `json:"runnerName"`
	OddValue        float64    `json:"oddValue"`
	StakeAmount     float64    `json:"stakeAmount"`
	SessionOutcome  float64    `json:"sessionOutcome"`
	OddsHistory     []OddsData `json:"oddsHistory"`
	SizeMatched     float64    `json:"sizeMatched"`
	SizeRemaining   float64    `json:"sizeRemaining"`
	MatchedOddValue float64    `json:"matchedOddValue"`
	IsUnmatched     bool       `json:"isUnmatched"`
}

type UpdateBetReqDto struct {
	Status           string    `json:"status"`
	ErrorDescription string    `json:"errorDescription"`
	BetUpdate        BetUpdate `json:"betUpdate"`
}
