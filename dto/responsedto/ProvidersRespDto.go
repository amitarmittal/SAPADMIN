package responsedto

type ProviderDto struct {
	ProviderName string `json:"providerName"`
	ProviderId   string `json:"providerId"`
	Status       string `json:"status"`
}

// GetSportsRespDto represents response body of this API
type ProvidersRespDto struct {
	Status           string        `json:"status"`
	ErrorDescription string        `json:"errorDescription"`
	Providers        []ProviderDto `json:"providers"`
}
