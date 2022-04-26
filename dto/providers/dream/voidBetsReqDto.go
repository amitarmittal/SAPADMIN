package dream

type VoidBetsReqDto struct {
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
	MarketId   string `json:"marketId"`
	MarketName string `json:"marketName"`
	MarketType string `json:"marketType"`
	StartTime  int64  `json:"startTime"` // unix-milli seconds
	EndTime    int64  `json:"endTime"`   // unix-milli seconds
	Reason     string `json:"reason"`
	//RollbackType string `json:"rollbackType"` // Rollback / Void
}
