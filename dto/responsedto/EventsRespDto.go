package responsedto

import dto "Sp/dto/core"

// GetEventsRespDto represents response body of this API
type EventsRespDto struct {
	Status           string         `json:"status"`
	ErrorDescription string         `json:"errorDescription"`
	Events           []dto.EventDto `json:"events"`
}
