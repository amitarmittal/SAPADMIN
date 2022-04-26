package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type CommissionReq struct {
	ReqId            string  `json:"reqId"`            // UniqueId
	TransactionId    string  `json:"transactionId"`    // UserMarketKey => OperatorId+"-"+UserId+"-"+MarketKey
	Token            string  `json:"token"`            // User's latest token
	OperatorId       string  `json:"operatorId"`       // Operator Id for that operator
	UserId           string  `json:"userId"`           // User Id for that user
	ProviderId       string  `json:"providerId"`       // BetFair
	SportId          string  `json:"sportId"`          //
	CompetitionId    string  `json:"competitionId"`    //
	EventId          string  `json:"eventId"`          //
	MarketId         string  `json:"marketId"`         // Unique Id of the market
	WinningAmount    float64 `json:"winningAmount"`    // Market Level Wining
	Commission       float64 `json:"commission"`       // Win Commission %
	CommissionAmount float64 `json:"commissionAmount"` // Actual Commission to Charge
	UserCommission   float64 `json:"userCommission"`   // Commission Charged on Wins
	CommissionCredit float64 `json:"commissionCredit"` // +ve - credit to the user, -ve - debit from the user
}
