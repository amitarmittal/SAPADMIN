package commondto

type OperStatusInCompetitionReqDto struct {
	OperatorId    string `json:"operatorId"`
	ProviderId    string `json:"providerId"`
	SportId       string `json:"sportId"`
	CompetitionId string `json:"competitionId"`
	Status        string `json:"status"`
}
