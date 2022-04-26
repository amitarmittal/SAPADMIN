package responsedto

type BetStatus struct {
	BetId  string `json:"betId"`
	Status string `json:"status"`
}

type BetsStatusRespDto struct {
	Status           string      `json:"status"`
	ErrorDescription string      `json:"errorDescription"`
	BetsStatus       []BetStatus `json:"betsStatus"`
}
