package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type DepositReqDto struct {
	OperatorId   string  `json:"operatorId"`
	PartnerId    string  `json:"partnerId"`
	UserId       string  `json:"userId"`
	CreditAmount float64 `json:"creditAmount"`
	Remark       string  `json:"remark"`
}
