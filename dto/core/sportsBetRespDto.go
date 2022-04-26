package dto

type SportsBetRespDto struct {
	Status           string       `json:"status"`
	ErrorDescription string       `json:"errorDescription"`
	Balance          float64      `json:"balance"`
	OpenBets         []OpenBetDto `json:"openBets"`
}

// default constructor
func NewSportsBetRespDto() SportsBetRespDto {
	return SportsBetRespDto{Status: "RS_ERROR", ErrorDescription: "Generic Error!", Balance: 0, OpenBets: []OpenBetDto{}}
}
