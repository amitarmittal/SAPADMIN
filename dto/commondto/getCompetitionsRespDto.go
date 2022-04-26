package commondto

type CompetitionDto struct {
	SportId         string `json:"sportId"`
	CompetitionName string `json:"competitionName"`
	CompetitionId   string `json:"competitionId"`
	MarketCount     int    `json:"marketCount"`
	Region          string `json:"region"`
}

// GetCompetitionsRespDto represents response body of this API
type GetCompetitionsRespDto struct {
	Status           string           `json:"status"`
	ErrorDescription string           `json:"errorDescription"`
	Competitions     []CompetitionDto `json:"competitions"`
}
