package reports

type UserBookRunner struct {
	RunnerName string  `json:"runnerName"`
	RunnerId   string  `json:"runnerId"`
	RunnerRisk float64 `json:"runnerRisk"`
}

type UserBook struct {
	UserId          string           `json:"userId"`
	UserName        string           `json:"userName"`
	UserBookRunners []UserBookRunner `json:"risk_report"`
}

type UserBookReportRespDto struct {
	Status           string     `json:"status"`
	ErrorDescription string     `json:"errorDescription"`
	EventId          string     `json:"eventId"`
	EventName        string     `json:"eventName"`
	MarketId         string     `json:"marketId"`
	MarketType       string     `json:"marketType"`
	MarketName       string     `json:"marketName"`
	UserBooks        []UserBook `json:"userBooks"`
}

type UserBookReportReqDto struct {
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
	MarketId   string `json:"marketId"`
}

type UserIdToRunnerRisk struct {
	UserId      string
	RunnerRisks map[string]float64
}
