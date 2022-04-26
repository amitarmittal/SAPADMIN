package dto

// GetSportsRespDto represents response body of this API
type GetUserBalanceRespDto struct {
	Status           string  `json:"status"`
	ErrorDescription string  `json:"errorDescription"`
	Balance          float64 `json:"balance"`
	Currency         string  `json:"currency"`
	WalletType       string  `json:"walletType"` // Seamless / Transfer
}
