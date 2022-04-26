package dream

type ResultReqDto struct {
	SportId        string  `json:"sportId"`
	EventId        string  `json:"eventId"`
	MarketId       string  `json:"marketId"`
	MarketName     string  `json:"marketName"`
	MarketType     string  `json:"marketType"`
	RunnerId       string  `json:"runnerId"`
	RunnerName     string  `json:"runnerName"`
	SessionOutcome float64 `json:"sessionOutcome"`
}
