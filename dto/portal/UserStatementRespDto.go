package dto

type UserTransaction struct {
	TxTime          int64   `json:"transactionTime"` // unix time in milliseconds
	TxType          string  `json:"transactionType"` //
	RefId           string  `json:"referenceId"`
	Amount          float64 `json:"amount"` // +ve value added to user, -ve values deducted from user
	Remark          string  `json:"remark"`
	CompetitionName string  `json:"competitionName"`
	EventName       string  `json:"eventName"`
	MarketType      string  `json:"marketType"`
	MarketName      string  `json:"marketName"`
}

type UserStatementRespDto struct {
	Status           string            `json:"status"`
	ErrorDescription string            `json:"errorDescription"`
	Transactions     []UserTransaction `json:"Transactions"`
	Balance          float64           `json:"balance"`
	Page             int               `json:"page"`         // Current Page number
	PageSize         int               `json:"pageSize"`     // Bets count in Bets Array
	TotalRecords     int               `json:"totalRecords"` // Total bet count which matched the filtered query
}
