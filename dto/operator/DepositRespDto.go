package operatordto

type DepositRespDto struct {
	Status           string  `json:"status"`
	ErrorDescription string  `json:"errorDescription"`
	Balance          float64 `json:"balance"`
}
