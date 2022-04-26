package requestdto

type BetsReportReq struct {
	Page       int64  `json:"page"`
	PageSize   int64  `json:"pageSize"`
	OperatorId string `json:"operatorId"`
}
