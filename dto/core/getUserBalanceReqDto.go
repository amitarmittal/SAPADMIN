package dto

//GetSportsReqDto
type GetUserBalanceReqDto struct {
	Token      string `json:"token"`
	OperatorId string `json:"operatorId"`
	PartnerId  string `json:"partnerId"`
}
