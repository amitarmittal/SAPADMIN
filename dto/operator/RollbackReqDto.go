package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type RollbackReqDto struct {
	OperatorId     string  `json:"operatorId"`
	UserId         string  `json:"userId"`
	Token          string  `json:"token"`
	ReqId          string  `json:"reqId"`
	TransactionId  string  `json:"transactionId"`
	EventId        string  `json:"eventId"`        // EventId map to gameId in JSON to be inline with GAP
	MarketId       string  `json:"marketId"`       // MarketId map to roundId in JSON to be inline iwth GAP
	RollbackAmount float64 `json:"rollbackAmount"` // +ve value to deposit to user, -ve value to deduct from user
	RollbackType   string  `json:"rollbackType"`   // rollback / void / cancelled / expired / lapsed / deleted
	RollbackReason string  `json:"rollbackReason"` // wrong settlement / market closed / invalid odds / time exceeded
}

/*
type RollbackReqDto struct {
	OperatorId     string  `json:"operatorId"`
	UserId         string  `json:"userId"`
	Token          string  `json:"token"`
	ReqId          string  `json:"reqId"`
	TransactionId  string  `json:"transactionId"`
	EventId        string  `json:"eventId"`
	MarketId       string  `json:"marketId"`
	RollbackAmount float64 `json:"rollbackAmount"`
}
*/
