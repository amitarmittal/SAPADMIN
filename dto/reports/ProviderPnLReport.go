package reports

type ProviderPnLReport struct {
	ProviderId string  `bson:"providerId"`
	SportName  string  `json:"sportName"`
	SportId    string  `json:"sportId"`
	ProfitLoss float64 `json:"profitLoss"`
	BetCount   int     `json:"betCount"`
}

type ProviderPnLReportRespDto struct {
	Status             string              `json:"status"`
	ErrorDescription   string              `json:"errorDescription"`
	ProviderId         string              `json:"providerId"`
	ProviderPnLReports []ProviderPnLReport `json:"providerPnLReport"`
}

type ProviderPnLReportReqDto struct {
	OperatorId string  `json:"operatorId"`
	ProviderId string  `json:"providerId"`
	UserId     string  `json:"userId"`
	StartTime  float64 `json:"startTime"`
	EndTime    float64 `json:"endTime"`
}
