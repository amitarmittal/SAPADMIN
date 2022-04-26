package operatordto

type BetDetailsReqDto struct {
	OperatorId string `json:"operatorId"`
	UserId     string `json:"userId"` // optional. Empty value to get all user bets
	BetId      string `json:"betId"`  // optional. Empty value to get bets across all providers
}
