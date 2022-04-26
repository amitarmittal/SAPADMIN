package reports

type Statement struct {
	TransactionTime int64   `json:"transactionTime"`
	TransactionId   string  `json:"transactionId"`
	DebitAmount     float64 `json:"debitAmount"`
	CreditAmount    float64 `json:"creditAmount"`
	Balance         float64 `json:"balance"`
	TransactionType string  `json:"transactionType"`
	ReferenceId     string  `json:"referenceId"`
	Remark          string  `json:"remark"`
}

type TransferUserStatementRespDto struct {
	Status           string      `json:"status"`
	ErrorDescription string      `json:"errorDescription"`
	UserId           string      `json:"userId"`
	UserName         string      `json:"userName"`
	UserBalance      float64     `json:"userBalance"`
	Statement        []Statement `json:"userStatement"`
}

type TransferUserStatementReqDto struct {
	UserId      string `json:"userId"`
	StartTime   int64  `json:"startTime"`
	EndTime     int64  `json:"endTime"`
	ReferenceId string `json:"referenceId"`
}
