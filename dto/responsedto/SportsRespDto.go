package responsedto

type SportDto struct {
	SportName string `json:"sportName"`
	SportId   string `json:"sportId"`
	Status    string `json:"status"`
	PartnerId string `json:"partnerId"`
}

// GetSportsRespDto represents response body of this API
type SportsRespDto struct {
	Status           string     `json:"status"`
	ErrorDescription string     `json:"errorDescription"`
	Sports           []SportDto `json:"sports"`
}
