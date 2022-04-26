package dto

type GetEventReqDto struct {
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"EventId"`
}
