package reports

type Runner struct {
	RunnerName string  `json:"runnerName"`
	RunnerId   string  `json:"runnerId"`
	RunnerRisk float64 `json:"runnerRisk"`
}

type Market struct {
	MarketId   string   `json:"marketId"`
	MarketType string   `json:"marketType"`
	MarketName string   `json:"marketName"`
	Runners    []Runner `json:"risk_report"`
}

type RiskReportRespDto struct {
	Status           string   `json:"status"`
	ErrorDescription string   `json:"errorDescription"`
	EventId          string   `json:"eventId"`
	EventName        string   `json:"eventName"`
	Markets          []Market `json:"markets"`
}

type RiskReportReqDto struct {
	ProviderId string `json:"providerId"`
	UserId     string `json:"userId"`
	EventId    string `json:"eventId"`
	// MarketIds  []string `json:"marketId"`
}

type MarketIdToRunnerRisk struct {
	MarketId    string
	RunnerRisks map[string]float64
}
