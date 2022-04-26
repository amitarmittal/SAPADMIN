package dto

type GetLiveEventsReqDto struct {
	Token      string `json:"token"`
	OperatorId string `json:"operatorId"`
	PartnerId  string `json:"partnerId"`
	ProviderId string `json:"providerId"`
}

//func NewGetSportsRequest(operatorId string, token string) *GetSportsReqDto {
//	return &GetSportsReqDto{OperatorId: operatorId, Token: token}
//}
