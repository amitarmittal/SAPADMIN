package responsedto

import dto "Sp/dto/core"

// GetEventsRespDto represents response body of this API
type MarketsRespDto struct {
	Status           string       `json:"status"`
	ErrorDescription string       `json:"errorDescription"`
	Event            dto.EventDto `json:"event"`
}
