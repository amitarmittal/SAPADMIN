package dto

type Event struct {
	EventId         string `json:"eventId"`
	ProviderId      string `json:"providerId"`
	SportsId        string `json:"sportsId"`
	SportName       string `json:"sportName"`
	CompetitionId   string `json:"competitionId"`
	CompetitionName string `json:"competitionName"`
	EventName       string `json:"eventName"`
	OpenDate        int64  `json:"openDate"` // Unix milliseconds
	Favourite       bool   `json:"favourite"`
	Status          string `json:"Status"`
}

type GetEventsRespDto struct {
	Status           string  `json:"status"`
	ErrorDescription string  `json:"errorDescription"`
	Events           []Event `json:"events"`
}

type BlockedEventResqDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
}

type UnblockedEventResqDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
}
