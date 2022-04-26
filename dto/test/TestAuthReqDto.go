package dto

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type TestAuthReqDto struct {
	OperatorId   string  `json:"operatorId"`
	OperatorName string  `json:"operatorName"`
	OperatorKey  string  `json:"operatorKey"`
	BaseURL      string  `json:"baseURL"`
	Status       string  `json:"status"`
	Currency     string  `json:"currency"`
	Commisssion  float64 `json:"commission"`
}
