package responsedto

type AllBetDto struct {
	BetId          string  `json:"betId"`
	BetType        string  `json:"betType"`
	BetStatus      string  `json:"betStatus"`
	OddValue       float64 `json:"oddValue"`
	StakeAmount    float64 `json:"stakeAmount"`
	RunnerName     string  `json:"runnerName"`
	RunnerId       string  `json:"runnerId"`
	MarketType     string  `json:"marketType"`
	MarketId       string  `json:"marketId"`
	MarketName     string  `json:"marketName"`
	EventId        string  `json:"eventId"`
	EventName      string  `json:"eventName"`
	SportId        string  `json:"sportId"`
	SportName      string  `json:"sportName"`
	SessionOutcome float64 `json:"sessionOutcome"`
	IsUnmatched    bool    `json:"isUnmatched"`
	RequestTime    int64   `json:"requestTime"`
	OpenEventDate  int64   `json:"openEventDate"`
	BetResult      string  `json:"betResult"`
	BetReturns     float64 `json:"betReturns"`
	Currency       string  `json:"currency"`
}

type GetAllBetsRespDto struct {
	Status           string      `json:"status"`
	ErrorDescription string      `json:"errorDescription"`
	AllBets          []AllBetDto `json:"allBets"`
}
