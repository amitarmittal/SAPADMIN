package reports

type PnLReport struct {
	SportName  string  `json:"sport_name"`
	MarketType string  `json:"market_type"`
	ProfitLoss float64 `json:"profit_loss"`
}

type PnLReportRespDto struct {
	Status           string      `json:"status"`
	ErrorDescription string      `json:"errorDescription"`
	UserName         string      `json:"user_name"`
	PnLReports       []PnLReport `json:"pnLReport"`
}

type PnLReportReqDto struct {
	UserName  string  `json:"user_name"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}
