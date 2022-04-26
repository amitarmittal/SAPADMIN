package dream

type RollbackReqDto struct {
	SportId      string `json:"sportId"`
	EventId      string `json:"eventId"`
	MarketId     string `json:"marketId"`
	MarketName   string `json:"marketName"`
	MarketType   string `json:"marketType"`
	RollbackType string `json:"rollbackType"` // Rollback / Void
	Reason       string `json:"reason"`
}
