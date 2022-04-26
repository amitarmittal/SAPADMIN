package reports

type SeemlessStatement struct {
	TransactionTime int64   `json:"transactionTime"`
	TransactionId   string  `json:"transactionId"`
	DebitAmount     float64 `json:"debitAmount"`
	CreditAmount    float64 `json:"creditAmount"`
	Balance         float64 `json:"balance"`
	TransactionType string  `json:"transactionType"`
	ReferenceId     string  `json:"referenceId"`
	Remark          string  `json:"remark"`
}

type SeemlessUserStatementRespDto struct {
	Status           string              `json:"status"`
	ErrorDescription string              `json:"errorDescription"`
	UserId           string              `json:"userId"`
	UserName         string              `json:"userName"`
	UserBalance      float64             `json:"userBalance"`
	Statement        []SeemlessStatement `json:"userStatement"`
}

type SeemlessUserStatementReqDto struct {
	UserId      string `json:"userId"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	ReferenceId string `json:"referenceId"`
}
