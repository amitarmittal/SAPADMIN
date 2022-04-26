package dto

type Competition struct {
	CompetitionId   string `json:"CompetitionId"`
	CompetitionName string `json:"CompetitionName"`
	Status          string `json:"status"` // active / blocked / deleted
}

type CompRespDto struct {
	Status           string        `json:"status"`
	ErrorDescription string        `json:"errorDescription"`
	Competitions     []Competition `json:"Competition"` //TODO: Add competitions DTO once it is created
}

type BlockedCompResqDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
}

type UnblockedCompResqDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
}
