package dto

type GetBalanceRespDto struct {
	Status           string  `json:"status"`
	ErrorDescription string  `json:"errorDescription"`
	Balance          float64 `json:"balance"`
	Currency         string  `json:"currency"`
}
