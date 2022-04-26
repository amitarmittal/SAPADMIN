package requestdto

type BetsStatusReqDto struct {
	OperatorId string   `json:"operatorId"`
	BetIds     []string `json:"betIds"`
}
