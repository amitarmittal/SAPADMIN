package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type BalanceReqDto struct {
	OperatorId string `json:"operatorId"`
	Token      string `json:"token"`
	UserId     string `json:"userId"`
}
