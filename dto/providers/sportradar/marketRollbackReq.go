package sportradar

/*
public String sportId;
public String eventId;
public String marketId;
public String marketName;
public String marketType;
public Date startTime;
public Date endTime;
public String reason;
public String rollbackType;
*/
type MarketRollbackReq struct {
	SportId      string `json:"sportId"`
	EventId      string `json:"eventId"`
	MarketId     string `json:"marketId"`
	MarketName   string `json:"marketName"`
	MarketType   string `json:"marketType"`
	StartTime    int64  `json:"startTime"`
	EndTime      int64  `json:"endTime"`
	Reason       string `json:"reason"`
	RollbackType string `json:"rollbackType"` // Rollback / TimelyVoid / TimelyVoidRollback
}
