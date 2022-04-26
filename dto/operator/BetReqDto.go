package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type BetReqDto struct {
	OperatorId    string  `json:"operatorId"`
	Token         string  `json:"token"`
	UserId        string  `json:"userId"`
	ReqId         string  `json:"reqId"`
	TransactionId string  `json:"transactionId"`
	BetType       string  `json:"betType"`
	OddValue      float64 `json:"oddValue"`
	EventId       string  `json:"eventId"`
	MarketId      string  `json:"marketId"`
	DebitAmount   float64 `json:"debitAmount"`
}
