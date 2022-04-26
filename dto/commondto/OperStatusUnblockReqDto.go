package commondto

type OperStatusUnblockReqDto struct {
	OperatorId    string `json:"operatorId"`
	PartnerId     string `json:"partnerId"`
	ProviderId    string `json:"providerId"`
	SportId       string `json:"sportId"`
	CompetitionId string `json:"competitionId"`
	EventId       string `json:"eventId"`
}
