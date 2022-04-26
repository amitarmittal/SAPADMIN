package commondto

type GetConfigReqDto struct {
	ConfigCall    string `json:"configCall"`
	OperatorId    string `json:"operatorId"`
	PartnerId     string `json:"partnerId"`
	ProviderId    string `json:"providerId"`
	SportId       string `json:"sportId"`
	CompetitionId string `json:"competitionId"`
	EventId       string `json:"eventId"`
}
