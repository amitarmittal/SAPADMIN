package dto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`
type CreateProviderReqDto struct {
	ProviderId     string `json:"providerId"`
	ProviderName   string `json:"providerName"`
	ProviderStatus string `json:"providerStatus"`
}
