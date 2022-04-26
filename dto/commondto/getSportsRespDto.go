package commondto

type SportDto struct {
	SportName string `json:"sportName"`
	SportId   string `json:"sportId"`
}

// GetSportsRespDto represents response body of this API
type GetSportsRespDto struct {
	Status           string     `json:"status"`
	ErrorDescription string     `json:"errorDescription"`
	Sports           []SportDto `json:"sports"`
}
