package dto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type TestAuthRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	Signature        string `json:"signature"`
}
