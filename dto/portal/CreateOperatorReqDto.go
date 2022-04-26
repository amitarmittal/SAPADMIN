package dto

import (
	"Sp/dto/commondto"
	operatordto "Sp/dto/operator"
)

// ProviderName string `json:"providerName"` // can be ENUM
// ProviderId string `json:"providerId"`

type CreateOperatorReqDto struct {
	OperatorId   string                `json:"operatorId"`
	OperatorName string                `json:"operatorName"`
	OperatorKey  string                `json:"operatorKey"`
	BaseURL      string                `json:"baseURL"`
	WalletType   string                `json:"walletType"`
	Status       string                `json:"status"`
	Partners     []operatordto.Partner `json:"partners"`
	Currency     string                `json:"currency"`
	Commisssion  float64               `json:"commission"`
	AddSubFields bool                  `json:"addSubFields"`
	Config       commondto.ConfigDto   `json:"config"`
	Ips          []string              `json:"ips"`
}

type PortalLoginReqDto struct {
	UserId   string `json:"userId"`
	Password string `json:"password"`
	IP       string `json:"ip"`
}

type GetDetailsReqDto struct {
	Page   int `json:"page"`
	Number int `json:"number"`
}
