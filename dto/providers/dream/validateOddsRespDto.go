package dream

// {"valid":false,"eventStatus":"IN_PLAY","eventDate":1631919600000,"oddsList":[1.91,1.9,1.89]}
type ValidateOddsRespDto struct {
	IsValid          bool      `json:"valid"`
	Status           string    `json:"eventStatus"`
	OpenDate         int64     `json:"eventDate"`
	OddValues        []float64 `json:"oddsList"`
	MatchedOddValue  float64   `json:"matchedOddValue"`
	ErrorDescription string    `json:"errorDescription"`
}
