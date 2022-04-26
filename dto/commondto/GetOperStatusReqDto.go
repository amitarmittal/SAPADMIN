package commondto

type GetOperDetailReqDto struct {
	ProviderId    string `json:"providerId"`
	SportId       string `json:"sportId"`
	CompetitionId string `json:"competitionId"`
	EventId       string `json:"eventId"`
}
