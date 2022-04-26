package dto

type OpenBetDto struct {
	BetId          string  `json:"betId"`
	BetType        string  `json:"betType"`
	OddValue       float64 `json:"oddValue"`
	StakeAmount    float64 `json:"stakeAmount"`
	RunnerName     string  `json:"runnerName"`
	MarketType     string  `json:"marketType"`
	SessionOutcome float64 `json:"sessionOutcome"`
}

// GetSportsRespDto represents response body of this API
type GetOpenBetsRespDto struct {
	Status           string       `json:"status"`
	ErrorDescription string       `json:"errorDescription"`
	OpenBets         []OpenBetDto `json:"openBets"`
}
