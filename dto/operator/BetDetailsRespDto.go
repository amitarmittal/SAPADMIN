package operatordto

/*
type RollbackReqDto struct {
	ReqId          string  `bson:"request_id"`
	ReqTime        int64   `bson:"request_time"`
	RollbackAmount float64 `bson:"rollback_amount"`
	RollbackReason string  `bson:"rollback_reason"`
}

type ResultReqDto struct {
	ReqId        string  `bson:"request_id"`
	ReqTime      int64   `bson:"request_time"`
	CreditAmount float64 `bson:"credit_amount"`
}

type BetReqDto struct {
	ReqId       string  `bson:"request_id"`
	ReqTime     int64   `bson:"request_time"`
	DebitAmount float64 `bson:"debit_amount"`
}

type BetDetailsDto struct {
	SportName       string  `bson:"sport_name"`
	CompetitionName string  `bson:"competition_name"`
	EventName       string  `bson:"event_name"`
	BetType         string  `bson:"betType"` // BACK/LAY
	OddValue        float64 `bson:"oddValue"`
	StakeAmount     int     `bson:"stakeAmount"`
	MarketType      string  `bson:"marketType"`     // MATCH_ODDS/BOOK_MAKER/GOAL_MARKETS/FANCY_MARKET/
	MarketName      string  `bson:"marketName"`     // for soccer, match_odds or goal_markets, can be avoided
	RunnerId        string  `bson:"runnerId"`       // ?? what is for fancy
	RunnerName      string  `bson:"runnerName"`     // ?? what is for fancy
	SessionOutcome  float64 `bson:"sessionOutcome"` // fancy scrore ex: 45 NO, 46 YES
}

type BetDto struct {
	EventKey     string           `bson:"event_key"`
	OperatorId   string           `bson:"operator_id"`
	Token        string           `bson:"token"`
	ProviderId   string           `bson:"provider_id"`
	UserId       string           `bson:"user_id"`
	EventId      string           `bson:"event_id"`
	MarketId     string           `bson:"market_id"`
	BetId        string           `bson:"transaction_id"`
	BetReq       BetReqDto        `bson:"bet_req"`
	ResultReqs   []ResultReqDto   `bson:"result_reqs"`
	RollbackReqs []RollbackReqDto `bson:"rollback_reqs"`
	BetDetails   BetDetailsDto    `bson:"bet_details"`
	Status       string           `bson:"status"` // unmatched / open(rollback) / settled / void / cancel / deleted
}
*/
type BetDetails struct {
	BetId string `json:"betId"`
}

//SportId    string  `json:"sportId"`
//EventId    string  `json:"eventId"`
//MarketId   string  `json:"marketId"`

type BetDetailsRespDto struct {
	Status           string       `json:"status"`
	ErrorDescription string       `json:"errorDescription"`
	Bets             []BetHistory `json:"bets"`
	Page             int          `json:"page"`     // optional. Empty value will bring latest results sort by date descending.
	PageSize         int          `json:"pageSize"` // optional. Empty value will bring 100 records. Value cant be more than 100.
}
