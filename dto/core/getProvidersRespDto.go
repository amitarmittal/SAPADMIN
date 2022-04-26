package dto

type ProviderInfo struct {
	ProviderName string `json:"providerName"`
	ProviderId   string `json:"providerId"`
	Status       string `json:"status"`
}

// GetSportsRespDto represents response body of this API
type GetProvidersRespDto struct {
	Status           string         `json:"status"`
	ErrorDescription string         `json:"errorDescription"`
	Providers        []ProviderInfo `json:"providers"`
}
