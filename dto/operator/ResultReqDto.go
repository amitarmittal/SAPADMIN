package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type ResultReqDto struct {
	OperatorId    string  `json:"operatorId"`
	Token         string  `json:"token"`
	UserId        string  `json:"userId"`
	ReqId         string  `json:"reqId"`
	TransactionId string  `json:"transactionId"`
	EventId       string  `json:"eventId"`  // EventId map to gameId in JSON to be inline with GAP
	MarketId      string  `json:"marketId"` // MarketId map to roundId in JSON to be inline iwth GAP
	CreditAmount  float64 `json:"creditAmount"`
}
