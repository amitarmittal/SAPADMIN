package dto

import (
	"Sp/dto/commondto"
)

type Operator struct {
	OperatorId   string              `json:"operator_id"`
	OperatorName string              `json:"operator_name"`
	BaseURL      string              `json:"base_url"`
	Balance      float64             `json:"balance"`
	Currency     string              `json:"currency"`
	Status       string              `json:"status"`      // ACTIVE / BLOCKED
	WalletType   string              `json:"wallet_type"` // Seamless / Transfer / Feed
	OperatorKey  string              `json:"operatorKey"`
	PublicKey    string              `json:"publicKey"`
	Config       commondto.ConfigDto `json:"config"`
}

type GetOperatorsRespDto struct {
	Status           string     `json:"status"`
	ErrorDescription string     `json:"errorDescription"`
	Operators        []Operator `json:"operators"`
}
