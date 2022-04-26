package dto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type AuthReqDto struct {
	OperatorId string `json:"operatorId"`
	PartnerId  string `json:"partnerId"`
	UserId     string `json:"userId"`
	PlatformId string `json:"platformId"` // can be ENUM
	Currency   string `json:"currency"`   // can be ENUM
	ClientIp   string `json:"clientIp"`
	Username   string `json:"username"`
}
