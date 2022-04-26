package responsedto

import dto "Sp/dto/core"

// GetEventsRespDto represents response body of this API
type LiveEventsRespDto struct {
	Status           string         `json:"status"`
	ErrorDescription string         `json:"errorDescription"`
	Events           []dto.EventDto `json:"events"`
}
