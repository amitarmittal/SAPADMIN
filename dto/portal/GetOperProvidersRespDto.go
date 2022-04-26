package dto

type OperProvider struct {
	ProviderId   string `json:"ProviderId"`
	ProviderName string `json:"ProviderName"`
	Status       string `json:"Status"` // active / blocked / deleted
}

type GetOperProvidersRespDto struct {
	Status           string         `json:"status"`
	ErrorDescription string         `json:"errorDescription"`
	OperProviders    []OperProvider `json:"operProviders"`
}
