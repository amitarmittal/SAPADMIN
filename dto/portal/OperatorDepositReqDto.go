package dto

type OperatorDepositReqDto struct {
	OperatorId   string  `json:"operatorId"`
	PartnerId    string  `json:"partnerId"`
	CreditAmount float64 `json:"creditAmount"`
	PassKey      string  `josn:"passKey"`
}
