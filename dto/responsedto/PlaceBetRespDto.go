package responsedto

type PlaceBetRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	BetId            string `json:"betId"`
}
