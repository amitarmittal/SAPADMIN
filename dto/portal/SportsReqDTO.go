package dto

type SportsReqDto struct {
	ProviderId string `json:"providerId"`
}

type BlockedSportReqDto struct {
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
}

type UnblockedSportReqDto struct {
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
}
