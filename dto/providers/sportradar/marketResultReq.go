package sportradar

type RunnerResult struct {
	RunnerId       string  `json:"runnerId"`
	RunnerName     string  `json:"runnerName"`
	Result         string  `json:"result"` // Won / Lost / UndecidedYet
	DeadHeatFactor float64 `json:"deadHeatFactor"`
	VoidFactor     float64 `json:"voidFactor"`
}

type Result struct {
	ReceivedAt      int64          `json:"receivedAt"`
	VoidReason      string         `json:"voidReason"`
	VoidReasonID    int            `json:"voidReasonId"`
	VoidDescription string         `json:"voidDescription"`
	Runners         []RunnerResult `json:"runners"`
}

type MarketResultReq struct {
	SportId      string `json:"sportId"`
	EventId      string `json:"eventId"`
	MarketId     string `json:"marketId"`
	MarketName   string `json:"marketName"`
	MarketType   string `json:"marketType"`
	MarketStatus string `json:"marketStatus"` // RESULT_SETTLEMENT / VOID_SETTLEMENT
	Result       Result `json:"result"`
}

/*
MarketSettlement
String sportId
String eventId
String marketId
String marketName
List<RunnerResult> runners;

RunnerResult -> class
String runnerId
String runnerName
OutcomeResult result
double deadHeatFactor
double voidFactor

OutcomeResult -> enum
Lost
Won
UndecidedYet
*/
