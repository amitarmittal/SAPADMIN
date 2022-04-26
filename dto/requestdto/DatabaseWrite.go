package requestdto

type DatabaseWriteReqDto struct {
	OperatorId string `json:"operatorId"`
	Message    string `json:"message"` // any message
}
