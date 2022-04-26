package dto

type CompReqDto struct {
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
}

type BlockedCompReqDto struct {
	ProviderId    string `json:"providerId"`
	SportId       string `json:"sportId"`
	CompetitionId string `json:"competitionId"`
}

type UnblockedCompReqDto struct {
	ProviderId    string `json:"providerId"`
	SportId       string `json:"sportId"`
	CompetitionId string `json:"competitionId"`
}
