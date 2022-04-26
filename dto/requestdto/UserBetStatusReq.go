package requestdto

type UserBetStatusReqDto struct {
	Token      string `json:"token"`
	OperatorId string `json:"operatorId"`
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
}
