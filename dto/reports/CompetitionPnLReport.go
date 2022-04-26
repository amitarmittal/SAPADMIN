package reports

type CompetitionPnLReport struct {
	CompetitionName string  `json:"competitionName"`
	CompetitionId   string  `json:"competitionId"`
	EventName       string  `json:"eventName"`
	EventId         string  `json:"eventId"`
	ProfitLoss      float64 `json:"profitLoss"`
	BetCount        int     `json:"betCount"`
}

type CompetitionPnLReportRespDto struct {
	Status                string                 `json:"status"`
	ErrorDescription      string                 `json:"errorDescription"`
	CompetitionId         string                 `json:"competitionId"`
	CompetitionPnLReports []CompetitionPnLReport `json:"competitionPnLReport"`
}

type CompetitionPnLReportReqDto struct {
	OperatorId    string  `json:"operatorId"`
	ProviderId    string  `json:"providerId"`
	SportId       string  `json:"sportId"`
	CompetitionId string  `json:"competitionId"`
	UserId        string  `json:"userId"`
	StartTime     float64 `json:"startTime"`
	EndTime       float64 `json:"endTime"`
}
