package reports

type AdminStatement struct {
	SettlementTime int64   `json:"settlement_time"`
	FromTo         string  `json:"from_to"`
	Points         float64 `json:"points"`
	Amount         float64 `json:"amount"`
	MyShare        float64 `json:"my_share"`
	Status         string  `json:"status"`
	TransactionId  string  `json:"transaction_id"`
}

type AdminStatementRespDto struct {
	Status           string           `json:"status"`
	ErrorDescription string           `json:"errorDescription"`
	AdminStatements  []AdminStatement `json:"adminStatement"`
}

type AdminStatementReqDto struct {
	UserName  string  `json:"user_name"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
	Status    string  `json:"status"`
}
