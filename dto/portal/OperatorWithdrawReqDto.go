package dto

type OperatorWithdrawReqDto struct {
	OperatorId  string  `json:"operatorId"`
	PartnerId   string  `json:"partnerId"`
	DebitAmount float64 `json:"debitAmount"`
	PassKey     string  `josn:"passKey"`
}
