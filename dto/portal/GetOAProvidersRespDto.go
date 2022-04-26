package dto

type Provider struct {
	PartnerId    string `json:"partnerId"`
	Currency     string `json:"currency"`
	Rate         int32  `json:"rate"`
	ProviderId   string `json:"ProviderId"`
	ProviderName string `json:"ProviderName"`
	Status       string `json:"Status"` // active / blocked / deleted
}

type GetOAProvidersRespDto struct {
	Status           string     `json:"status"`
	ErrorDescription string     `json:"errorDescription"`
	Providers        []Provider `json:"providers"`
}
