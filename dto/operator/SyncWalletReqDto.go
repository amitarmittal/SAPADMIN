package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type SyncWalletReqDto struct {
	UserId  string  `json:"userId"`
	Balance float64 `json:"amount"`
}
