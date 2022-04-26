package responsedto

type ValidateOddsRespDto struct {
	Status           string    `json:"status"`
	ErrorDescription string    `json:"errorDescription"`
	IsValid          bool      `json:"isValid"`
	EventStatus      string    `json:"eventStatus"`
	EventDate        int64     `json:"eventDate"`
	OddValues        []float64 `json:"oddsList"`
	MatchedOddValue  float64   `json:"matchedOddValue"`
}
