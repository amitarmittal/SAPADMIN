package requestdto

type OpenBetsReqDto struct {
	Token      string `json:"token"`
	OperatorId string `json:"operatorId"` // Optional for feed-service
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
}
