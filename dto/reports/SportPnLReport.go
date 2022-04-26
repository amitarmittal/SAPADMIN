package reports

type SportPnLReport struct {
	SportName       string  `json:"sportName"`
	SportId         string  `json:"sportId"`
	CompetitionName string  `json:"competitionName"`
	CompetitionId   string  `json:"competitionId"`
	ProfitLoss      float64 `json:"profitLoss"`
	BetCount        int     `json:"betCount"`
}

type SportPnLReportRespDto struct {
	Status           string           `json:"status"`
	ErrorDescription string           `json:"errorDescription"`
	SportId          string           `json:"sportId"`
	SportPnLReports  []SportPnLReport `json:"sportPnLReport"`
}

type SportPnLReportReqDto struct {
	OperatorId string  `json:"operatorId"`
	ProviderId string  `json:"providerId"`
	SportId    string  `json:"sportId"`
	UserId     string  `json:"userId"`
	StartTime  float64 `json:"startTime"`
	EndTime    float64 `json:"endTime"`
}
