package responsedto

type OddsData struct {
	OddsKey   string  `json:"oddsKey"` // BEFORE / CURRENT / AFTER
	OddsValue float64 `json:"oddsValue"`
	OddsAt    string  `json:"oddsAt"` // Date in readable string format in UTC (from UNIX timestamp)
}

type OpenBetDto struct {
	BetId          string     `json:"betId"`
	BetType        string     `json:"betType"`
	OddValue       float64    `json:"oddValue"`
	StakeAmount    float64    `json:"stakeAmount"`
	RunnerName     string     `json:"runnerName"`
	RunnerId       string     `json:"runnerId"`
	MarketType     string     `json:"marketType"`
	MarketName     string     `json:"marketName"`
	MarketId       string     `json:"marketId"`
	EventId        string     `json:"eventId"`
	SportId        string     `json:"sportId"`
	SessionOutcome float64    `json:"sessionOutcome"`
	IsUnmatched    bool       `json:"isUnmatched"`
	OddsHistory    []OddsData `json:"oddsHistory"`
}

type OpenBetsRespDto struct {
	Status           string       `json:"status"`
	ErrorDescription string       `json:"errorDescription"`
	OpenBets         []OpenBetDto `json:"openBets"`
}
