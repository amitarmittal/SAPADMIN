package dto

type GetOperatorDetailsRespDto struct {
	Status           string     `json:"status"`
	ErrorDescription string     `json:"errorDescription"`
	Operator         []Operator `json:"operator"`
}
