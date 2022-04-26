package dto

type OperatorUser struct {
	UserId   string  `json:"userId"`
	UserName string  `json:"userName"`
	Balance  float64 `json:"balance"`
	Status   string  `json:"status"` // active / blocked / deleted
}

type GetUsersRespDto struct {
	Status           string         `json:"status"`
	ErrorDescription string         `json:"errorDescription"`
	Users            []OperatorUser `json:"users"`
	Page             int            `json:"page"`         // Current Page number
	PageSize         int            `json:"pageSize"`     // Bets count in Bets Array
	TotalRecords     int            `json:"totalRecords"` // Total bet count which matched the filtered query
}
