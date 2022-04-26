package operatordto

type GetBetRespDto struct {
	Status           string     `json:"status"`           // Request Status. "RS_OK" for Success, "RS_ERROR" for Failure
	ErrorDescription string     `json:"errorDescription"` // Failure reason
	BetDetails       BetHistory `json:"betDetails"`       // Bet Details
}
