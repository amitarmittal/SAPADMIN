package dto

type GetEventsReqDto struct {
	ProviderId    string `json:"providerId"`
	SportId       string `json:"sportId"`
	EventId       string `json:"EventId"`
	CompetitionId string `json:"competitionId"`
}

type BlockedEventReqDto struct {
	ProviderId    string `json:"providerId"`
	SportId       string `json:"sportId"`
	CompetitionId string `json:"competitionId"`
	EventId       string `json:"eventId"`
}

type UnblockedEventReqDto struct {
	ProviderId    string `json:"providerId"`
	SportId       string `json:"sportId"`
	CompetitionId string `json:"competitionId"`
	EventId       string `json:"eventId"`
}
