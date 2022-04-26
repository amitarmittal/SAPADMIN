package dto

type OpenBetDto struct {
	BetId          string  `json:"betId"`
	BetType        string  `json:"betType"`
	OddValue       float64 `json:"oddValue"`
	StakeAmount    float64 `json:"stakeAmount"`
	RunnerName     string  `json:"runnerName"`
	RunnerId       string  `json:"runnerId"`
	MarketType     string  `json:"marketType"`
	MarketName     string  `json:"marketName"`
	MarketId       string  `json:"marketId"`
	EventId        string  `json:"eventId"`
	SportId        string  `json:"sportId"`
	SessionOutcome float64 `json:"sessionOutcome"`
	IsUnmatched    bool    `json:"isUnmatched"`
	UserId         string  `json:"userId"`
	OperatorId     string  `json:"operatorId"`
	BetTime        int64   `json:"betTime"`
}

type OpenBetsRespDto struct {
	Status           string       `json:"status"`
	ErrorDescription string       `json:"errorDescription"`
	OpenBets         []OpenBetDto `json:"openBets"`
}
