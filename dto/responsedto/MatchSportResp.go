package responsedto

type MatchedSport struct {
	ID                string `json:"id"`
	SportID           string `json:"sportId"`
	CompetitionID     string `json:"competitionId"`
	CompetitionName   string `json:"competitionName"`
	EventName         string `json:"eventName"`
	EventID           string `json:"eventId"`
	SrSportID         string `json:"srSportId"`
	SrCompetitionID   string `json:"srCompetitionId"`
	SrCompetitionName string `json:"srCompetitionName"`
	SrEventName       string `json:"srEventName"`
	SrEventID         string `json:"srEventId"`
	OpenDate          int64  `json:"openDate"`
}

type MatchSportResp struct {
	Status           string         `json:"status"`
	ErrorDescription string         `json:"errorDescription"`
	MatchedSports    []MatchedSport `json:"matchedSports"`
}
