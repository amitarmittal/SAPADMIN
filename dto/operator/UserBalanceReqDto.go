package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type UserBalanceReqDto struct {
	OperatorId string `json:"operatorId"`
	PartnerId  string `json:"partnerId"`
	UserId     string `json:"userId"`
}
