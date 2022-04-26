package operatordto

type BetsHistoryReqDto struct {
	OperatorId string   `json:"operatorId"`     // Mandatory.
	UserId     string   `json:"userId"`         // optional. Empty value to get all user bets
	ProviderId string   `json:"providerId"`     // optional. Empty value to get bets across all providers
	SportId    string   `json:"sportId"`        // optional. Empty value to get bets across all sports
	EventId    string   `json:"eventId"`        // optional. Empty value to get bets across all sports
	BetIds     []string `json:"transactionIds"` // optional. Empty value to get bets across all sports
	Status     string   `json:"status"`         // optional. Empty value to get all. // OPEN / SETTLED / VOID / UNMATCHED / CANCEL / DELETED
	StartDate  int64    `json:"startDate"`      // optional. Date in unix format.
	EndDate    int64    `json:"endDate"`        // optional. Date in unix format.
	FilterBy   string   `json:"filterBy"`       // optional. Posible values are 'CreatedAt' and 'UpdatedAt'. Default is 'CreatedAt'
	Page       int      `json:"page"`           // optional. Empty value will bring latest results sort by date descending.
	PageSize   int      `json:"pageSize"`       // optional. Empty value will bring 50 records. Value can't be more than 50.
}
