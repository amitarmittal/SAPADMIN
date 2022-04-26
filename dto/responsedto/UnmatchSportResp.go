package responsedto

type UnmatchedSport struct {
	ID              string `json:"id"`
	SportID         string `json:"sportId"`
	CompetitionID   string `json:"competitionId"`
	CompetitionName string `json:"competitionName"`
	EventName       string `json:"eventName"`
	EventID         string `json:"eventId"`
	OpenDate        int64  `json:"openDate"`
}

type UnmatchSportResp struct {
	Status           string           `json:"status"`
	ErrorDescription string           `json:"errorDescription"`
	UnmatchedSports  []UnmatchedSport `json:"matchedSports"`
}
