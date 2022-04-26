package dto

type OpenBetsReqDto struct {
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
}
