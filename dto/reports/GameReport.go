package reports

type GameReport struct {
	GameName  string  `json:"game_name"`
	BetCount  int64   `json:"bet_count"`
	WinCount  int64   `json:"win_count"`
	LoseCount int64   `json:"lose_count"`
	VoidCount int64   `json:"void_count"`
	WinAmount float64 `json:"win_amount"`
}

type GameReportRespDto struct {
	Status           string       `json:"status"`
	ErrorDescription string       `json:"errorDescription"`
	UserName         string       `json:"user_name"`
	GameReports      []GameReport `json:"gameReport"`
}

type GameReportReqDto struct {
	UserName   string  `json:"user_name"`
	ProviderId string  `json:"provider_id"`
	StartTime  float64 `json:"start_time"`
	EndTime    float64 `json:"end_time"`
	Status     string  `json:"status"`
}
