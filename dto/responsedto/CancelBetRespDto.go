package responsedto

type CancelBetResp struct {
	Status        string  `json:"status"`
	ErrorCode     string  `json:"errorCode"`
	SizeCancelled float64 `json:"sizeCancelled"`
	SizeMatched   float64 `json:"sizeMatched"`
	SizeRemaining float64 `json:"sizeRemaining"`
	BetId         string  `json:"betId"`
	//BetId         string  `josn:"betId"`
}
type CancelBetRespDto struct {
	Status           string          `json:"status"`
	ErrorDescription string          `json:"errorDescription"`
	Balance          float64         `json:"balance"`
	CancelBetsResp   []CancelBetResp `json:"cancelBetsResp"`
	//OpenBets       []OpenBetDto    `json:"openBets"`
}
