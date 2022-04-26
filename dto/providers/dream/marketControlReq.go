package dream

type MarketControlReq struct {
	// SportId      string `json:"sportId"`
	EventId    string `json:"eventId"`
	MarketId   string `json:"marketId"`
	MarketType string `json:"marketType"`
	SessionId  string `json:"sessionId"` // Rollback / Void
}
