package dto

type GetBetReqDto struct {
	OperatorId string `json:"operatorId"`    // mandatory
	TxId       string `json:"transactionId"` // mandatory
}
