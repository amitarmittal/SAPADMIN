package dto

type UserStatementReqDto struct {
	OperatorId string `json:"operatorId"`      // mandatory
	UserId     string `json:"userId"`          // mandatory
	TxType     string `json:"transactionType"` // optional. Valid values are FUNDS / BETS
	StartDate  int64  `json:"startDate"`       // optional. Date in unix format.
	EndDate    int64  `json:"endDate"`         // optional. Date in unix format.
	Page       int    `json:"page"`            // optional. Empty value will bring latest results sort by date descending.
	PageSize   int    `json:"pageSize"`        // optional. Empty value will bring 50 records. Value can't be more than 50.
}
