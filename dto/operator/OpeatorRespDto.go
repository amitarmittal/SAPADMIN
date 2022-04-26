package operatordto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type OperatorRespDto struct {
	Balance float64 `json:"balance"`
	Status  string  `json:"status"`
}

type OperatorResultRespDto struct {
	Balance float64 `json:"balance"`
	Status  string  `json:"status"`
}

type OperatorRollbackRespDto struct {
	Balance float64 `json:"balance"`
	Status  string  `json:"status"`
}
