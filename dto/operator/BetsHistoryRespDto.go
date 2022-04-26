package operatordto

type BetRollback struct {
	RollbackReason string  `json:"rollbackReason"` // Rollback Reason
	RollbackTime   int64   `json:"rollbackTime"`   // Rollback anounced time
	RollbackAmount float64 `json:"rollbackAmount"` // Roolback amount. +ve value to add to user balance, -ve value to deduct from user balance
}
type BetResult struct {
	RunnerName     string  `json:"runnerName"`     // Actual outcome for MATCH_ODDS and BOOKMAKER markets, Market Name for FANCY markets
	SessionOutcome float64 `json:"sessionOutcome"` // Actual outcome for FANCY
	ResultTime     int64   `json:"resultTime"`     // Result announced time
	WinAmount      float64 `json:"winAmount"`      // User win amount
}

type BetReq struct {
	BetType        string  `json:"betType"`        // BACK or LAY bet
	BetOdds        float64 `json:"betOdds"`        // Odd value. Ex: 1.86
	UnifiedOdds    float64 `json:"unifiedOdds"`    // MatchOdds style odds
	BetStake       float64 `json:"betStake"`       // Bet Stake amount entered in the bet slip
	RunnerName     string  `json:"runnerName"`     // Expected outcome for MATCH_ODDS and BOOKMAKER markets, Market Name for FANCY markets
	SessionOutcome float64 `json:"sessionOutcome"` // Expected outcome for FANCY
	BetTime        int64   `json:"betTime"`        // Bet Placement Time
}

type BetHistory struct {
	BetId           string        `json:"betId"`           // Unique Bet Id
	UserId          string        `json:"userId"`          // Unique to Operator
	UserName        string        `json:"userName"`        // User's Friendly name to display
	ProviderName    string        `json:"providerName"`    // Sports provider name to display
	SportName       string        `json:"sportName"`       // Sport Name to display
	CompetitionName string        `json:"competitionName"` // Competition / League name to display
	EventId         string        `json:"eventId"`         // Event Id
	EventName       string        `json:"eventName"`       // Event / Match name to display
	MarketType      string        `json:"marketType"`      // Market Type
	MarketName      string        `json:"marketName"`      // Market Name to display
	BetAmount       float64       `json:"betAmount"`       // Debit Amount
	WonAmount       float64       `json:"wonAmount"`       // CreditAmount + RollbackAmount - DebitAmount. A positive number is a user win.
	BetTime         int64         `josn:"betTime"`         // Bet Placement Time
	Status          string        `json:"status"`          // Bet status. Open / Settled / Void
	UserIp          string        `json:"userIp"`          // User's device IP
	UpdatedAt       int64         `json:"updatedAt"`       // Last update time in unix format
	BetReq          BetReq        `json:"betReq"`          // Bet Request information. Refer inner details
	BetResults      []BetResult   `json:"betResults"`      // Array of Bet Results information. Refer inner details
	BetRollbacks    []BetRollback `json:"betRollbacks"`    // Array of Bet Rollback information. Refer inner details
}

type BetsHistoryRespDto struct {
	Status           string       `json:"status"`           // Request Status. "RS_OK" for Success, "RS_ERROR" for Failure
	ErrorDescription string       `json:"errorDescription"` // Failure reason
	Bets             []BetHistory `json:"bets"`             // Array of Bets based on filtered query
	Page             int          `json:"page"`             // Current Page number
	PageSize         int          `json:"pageSize"`         // Bets count in Bets Array
	TotalRecords     int          `json:"totalRecords"`     // Total bet count which matched the filtered query
}
