package reports

type EventPnLReport struct {
	EventId    string  `json:"eventId"`
	EventName  string  `json:"eventName"`
	MarketType string  `json:"marketType"`
	MarketName string  `json:"marketName"`
	MarketId   string  `json:"marketId"`
	ProfitLoss float64 `json:"profitLoss"`
	BetCount   int     `json:"betCount"`
}

type EventPnLReportRespDto struct {
	Status           string           `json:"status"`
	ErrorDescription string           `json:"errorDescription"`
	EventId          string           `json:"eventId"`
	EventPnLReports  []EventPnLReport `json:"sportPnLReport"`
}

type EventPnLReportReqDto struct {
	OperatorId    string  `json:"operatorId"`
	ProviderId    string  `json:"providerId"`
	SportId       string  `json:"sportId"`
	EventId       string  `json:"eventId"`
	CompetitionId string  `json:"competitionId"`
	UserId        string  `json:"userId"`
	StartTime     float64 `json:"startTime"`
	EndTime       float64 `json:"endTime"`
}
