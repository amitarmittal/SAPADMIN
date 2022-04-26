package reports

type UserStatement struct {
	SettlementTime int64   `json:"settlement_time"`
	EventName      string  `json:"event_name"`
	MarketType     string  `json:"market_type"`
	CreditAmount   float64 `json:"credit_amount"`
	DebitAmount    float64 `json:"debit_amount"`
	Balance        float64 `json:"balance"`
	TransactionId  string  `json:"transaction_id"`
}

type UserStatementRespDto struct {
	Status           string          `json:"status"`
	ErrorDescription string          `json:"errorDescription"`
	UserStatements   []UserStatement `json:"userStatement"`
}

type UserStatementReqDto struct {
	UserName  string  `json:"user_name"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Status    string  `json:"status"`
}
