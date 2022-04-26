package responsedto

type ValidateTokenRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	IsValid          bool   `json:"isValid"`
}
