package dto

type GetEventRespDto struct {
	Status           string `json:"status"`
	ErrorDescription string `json:"errorDescription"`
	Event            Event  `json:"events"`
}
