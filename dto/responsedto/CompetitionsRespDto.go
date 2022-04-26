package responsedto

type CompetitionDto struct {
	SportName       string `json:"sportName"`
	SportId         string `json:"sportId"`
	CompetitionName string `json:"competitionName"`
	CompetitionId   string `json:"competitionId"`
	Status          string `json:"status"`
	PartnerId       string `json:"partnerId"`
}

// GetSportsRespDto represents response body of this API
type CompetitionsRespDto struct {
	Status           string           `json:"status"`
	ErrorDescription string           `json:"errorDescription"`
	Competitions     []CompetitionDto `json:"competitions"`
}
