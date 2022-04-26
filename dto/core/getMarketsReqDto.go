package dto

type GetMarketsReqDto struct {
	Token      string `json:"token"`
	OperatorId string `json:"operatorId"`
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
}

//func NewGetSportsRequest(operatorId string, token string) *GetSportsReqDto {
//	return &GetSportsReqDto{OperatorId: operatorId, Token: token}
//}
