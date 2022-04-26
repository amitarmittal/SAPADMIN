package dto

//GetSportsReqDto
type AddUserCreditsReqDto struct {
	Token        string  `json:"token"`
	OperatorId   string  `json:"operatorId"`
	PartnerId    string  `json:"partnerId"`
	CreditAmount float64 `json:"creditAmount"`
}
