package dto

// ProviderId       string `json:"providerId"`
// Providers
type ProviderDto struct {
	ProviderId   string `json:"providerId"`
	ProviderName string `json:"providerName"`
	Status       string `json:"status"`
	Url          string `json:"url"`
}

// AuthRespDto represents response body of this API
type AuthRespDto struct {
	Status           string        `json:"status"`
	ErrorDescription string        `json:"errorDescription"`
	UserId           string        `json:"userId"`
	Token            string        `json:"token"`
	Url              string        `json:"url"`
	Providers        []ProviderDto `json:"providers"`
}
