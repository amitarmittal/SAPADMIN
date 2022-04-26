package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type UserBalanceRespDto struct {
	Status           string  `json:"status"`
	ErrorDescription string  `json:"errorDescription"`
	Balance          float64 `json:"balance"`
}
