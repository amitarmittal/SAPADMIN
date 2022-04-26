package dto

// GetEventsReqDto represents request body of this API
type GetEventsReqDto struct {
	Token      string `json:"token"`
	OperatorId string `json:"operatorId"`
	PartnerId  string `json:"partnerId"`
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
}

//func NewGetSportsRequest(operatorId string, token string) *GetSportsReqDto {
//	return &GetSportsReqDto{OperatorId: operatorId, Token: token}
//}
