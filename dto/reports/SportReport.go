package reports

type MarketStatement struct {
	BetTime        int64   `json:"betTime"`
	BetType        string  `json:"betType"`
	Currency       string  `json:"currency"`
	Rate           int32   `json:"rate"`
	EventName      string  `json:"eventName"`
	MarketName     string  `json:"marketName"`
	MarketType     string  `json:"marketType"`
	OddValue       float64 `json:"odds"`
	RunnerName     string  `json:"runnerName"`
	SessionOutCome float64 `json:"sessionOutcome"`
	StakeAmount    float64 `json:"stake"`
	Returns        float64 `json:"netAmount"`
	TransactionId  string  `json:"transactionId"`
	UserName       string  `json:"userName"`
	UserId         string  `json:"userId"`
	SportName      string  `json:"sportName"`
	SportId        string  `json:"sportId"`
	Status         string  `json:"status"`
}

type SportReport struct {
	EventName  string            `json:"eventName"`
	EventId    string            `json:"eventId"`
	ProviderId string            `json:"providerId"`
	BetCount   int64             `json:"betCount"`
	NetAmount  float64           `json:"netAmount"`
	Bets       []MarketStatement `json:"bets"`
}

type SportReportRespDto struct {
	Status           string        `json:"status"`
	ErrorDescription string        `json:"errorDescription"`
	UserId           string        `json:"userId"`
	SportReports     []SportReport `json:"sportReport"`
}

type SportReportReqDto struct {
	UserName      string  `json:"userName"`
	UserId        string  `json:"userId"`
	ProviderId    string  `json:"providerId"`
	EventId       string  `json:"eventId"`
	CompetitionId string  `json:"competitionId"`
	SportId       string  `json:"sportId"`
	MarketId      string  `json:"marketId"`
	StartTime     float64 `json:"startTime"`
	EndTime       float64 `json:"endTime"`
	Status        string  `json:"status"`
}
