package dto

// GetSportsRespDto represents response body of this API
type AddUserCreditsRespDto struct {
	Status           string  `json:"status"`
	ErrorDescription string  `json:"errorDescription"`
	Balance          float64 `json:"balance"`
	WalletType       string  `json:"walletType"` // Seamless / Transfer
}
