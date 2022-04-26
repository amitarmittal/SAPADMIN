package requestdto

type BetsResultReqDto struct {
	OperatorId string   `json:"operatorId"`
	BetIds     []string `json:"betIds"`
}
