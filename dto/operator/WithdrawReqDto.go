package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type WithdrawReqDto struct {
	OperatorId  string  `json:"operatorId"`
	PartnerId   string  `json:"partnerId"`
	UserId      string  `json:"userId"`
	DebitAmount float64 `json:"debitAmount"`
	Remark      string  `json:"remark"`
}
