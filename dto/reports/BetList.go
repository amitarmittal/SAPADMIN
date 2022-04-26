package reports

type BetList struct {
	BetTime        int64   `json:"betTime"`
	UserName       string  `json:"userName"`
	SportName      string  `json:"sportName"`
	EventName      string  `json:"eventName"`
	MarketType     string  `json:"marketType"`
	MarketName     string  `json:"marketName"`
	RunnerName     string  `json:"runnerName"`
	PartnerId      string  `json:"partnerId"`
	Currency       string  `json:"currency"`
	Rate           int32   `json:"rate"`
	Odds           float64 `json:"odds"`
	BetType        string  `json:"betType"`
	Stake          float64 `json:"stake"`
	SessionOutCome float64 `json:"sessionOutcome"`
	TransactionId  string  `json:"transactionId"`
	Status         string  `json:"status"`
	OperatorHold   float64 `json:"operatorHold"`
	NetAmount      float64 `json:"netAmount"`
	OperatorAmount float64 `json:"operatorAmount"`
	OperatorId     string  `json:"operatorId"`
}

type BetListRespDto struct {
	Status           string    `json:"status"`
	ErrorDescription string    `json:"errorDescription"`
	BetLists         []BetList `json:"betList"`
}

type BetListReqDto struct {
	UserId     string  `json:"userId"`
	StartTime  float64 `json:"startTime"`
	EndTime    float64 `json:"endTime"`
	SportName  string  `json:"sportName"`
	SportId    string  `json:"sportId"`
	Status     string  `json:"status"`
	ProviderId string  `json:"providerId"`
	OperatorId string  `json:"operatorId"`
}
