package reports

import "Sp/dto/models"

type BetReportDetail struct {
	BetId        string        `json:"betId"`
	BetStatus    string        `json:"betStatus"`
	ResultTime   int64         `json:"resultTime"`
	LapsedTime   int64         `json:"lapsedTime"`
	LapsedAmount float64       `json:"lapsedAmount"`
	BetType      string        `json:"betType"`
	Odds         float64       `json:"odds"`
	MarketResult models.Result `json:"marketResult"`
}

type BetDetailReportRespDto struct {
	Status           string          `json:"status"`
	ErrorDescription string          `json:"errorDescription"`
	BetReportDetail  BetReportDetail `json:"betReportDetail"`
}

type BetDetailReportReqDto struct {
	TransactionId string `json:"transactionId"`
}
