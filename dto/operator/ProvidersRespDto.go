package operatordto

import dto "Sp/dto/core"

// GetSportsRespDto represents response body of this API
type GetProvidersRespDto struct {
	Status           string             `json:"status"`
	ErrorDescription string             `json:"errorDescription"`
	Providers        []dto.ProviderInfo `json:"providers"`
}
