package dto

type Sport struct {
	SportId   string `json:"SportId"`
	SportName string `json:"SportName"`
	Status    string `json:"status"` // active / blocked / deleted
	PartnerId string `json:"partnerId"`
}

type SportsRespDto struct {
	Status           string  `json:"status"`
	ErrorDescription string  `json:"errorDescription"`
	Sports           []Sport `json:"sports"`
}

type BlockedSportRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
}

type UnblockedSportRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
}
