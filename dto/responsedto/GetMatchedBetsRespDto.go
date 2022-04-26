package responsedto

type IsMatchedStatus struct {
	BetId         string  `json:"betId"`
	IsMatched     bool    `json:"isMatched"`
	BetStatus     string  `json:"betStatus"`
	OddValue      float64 `json:"oddValue"`
	SizeMatched   float64 `json:"sizeMatched"`
	SizeRemaining float64 `json:"sizeRemaining"`
}

type GetMatchedBetsRespDto struct {
	Status           string            `json:"status"`
	ErrorDescription string            `json:"errorDescription"`
	IsMatchedStatus  []IsMatchedStatus `json:"betIds"`
}
