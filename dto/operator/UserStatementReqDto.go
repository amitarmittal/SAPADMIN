package operatordto

type UserStatementReqDto struct {
	OperatorId  string `json:"operatorId"`
	UserId      string `json:"userId"`
	Token       string `json:"token"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	ReferenceId string `json:"referenceId"`
}
