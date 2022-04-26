package dto

//GetSportsReqDto
type GetOpenBetsReqDto struct {
	Token      string `json:"token"`
	OperatorId string `json:"operatorId"`
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
}
