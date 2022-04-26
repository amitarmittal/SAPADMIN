package responsedto

import (
	sportdto "Sp/dto/sports"
)

type TotalCount struct {
	Count int64 `bson:"count"`
}
type BetData struct {
	Data  []sportdto.BetDto `json:"data" bson:"data"`
	Total []TotalCount      `json:"total" bson:"total"`
}
type BetReportResp struct {
	Status           string            `json:"status"`
	ErrorDescription string            `json:"errorDescription"`
	Page             int64             `json:"page"`
	PageSize         int64             `json:"pageSize"`
	TotalRecords     int64             `json:"totalRecords"`
	BetsData         []sportdto.BetDto `json:"betsData"`
}
