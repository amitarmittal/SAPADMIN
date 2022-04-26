package requestdto

type AddSportReqDto struct {
	ProviderId string `json:"providerId"` // DREAM/BET_FAIR/SPORTS_RADAR
	SportId    string `json:"sportId"`    // CRICKET/SOCCER/TENNIS
	SportName  string `json:"sportName"`
	Status     string `json:"status"`
}
