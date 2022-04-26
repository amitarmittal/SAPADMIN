package dto

type Partner struct {
	PartnerId    string  `json:"partner_id"`
	OperatorId   string  `json:"operator_id"`
	OperatorName string  `json:"operator_name"`
	Currency     string  `json:"currency"`
	Rate         int32   `json:"rate"`
	Commission   float64 `json:"commission"`
	Status       string  `json:"status"`      // ACTIVE / BLOCKED / SUSPENDED / DEACTIVATED
	WalletType   string  `json:"wallet_type"` // Seamless / Transfer
}

type GetPartnersRespDto struct {
	Status           string    `json:"status"`
	ErrorDescription string    `json:"errorDescription"`
	Partners         []Partner `json:"partners"`
}
