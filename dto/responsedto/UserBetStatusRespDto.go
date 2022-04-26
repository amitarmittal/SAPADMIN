package responsedto

type UserBetStatusRespDto struct {
	Status           string       `json:"status"`           // RS_OK / RS_ERROR
	ErrorDescription string       `json:"errorDescription"` // ERROR Description if RS_ERROR
	BetStatus        string       `json:"betStatus"`        // PENDING / COMPLETED / CANCELLED / EXPIRED
	BetErrorMsg      string       `json:"betErrorMsg"`      // bet placement failure message
	BetReqTime       int64        `json:"requestTime"`      // Last BET request time
	BetId            string       `json:"betId"`            // Last Bet Id
	Balance          float64      `json:"balance"`
	OpenBets         []OpenBetDto `json:"openBets"`
}
