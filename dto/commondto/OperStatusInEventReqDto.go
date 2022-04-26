package commondto

type OperStatusInEventReqDto struct {
	OperatorId string `json:"operatorId"`
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
	Status     string `json:"status"`
}
