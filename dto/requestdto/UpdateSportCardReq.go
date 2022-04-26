package requestdto

type UpdateSportCardReq struct {
	EventID           string `json:"eventId"`
	SrCompetitionID   string `json:"srCompetitionId"`
	SrCompetitionName string `json:"srCompetitionName"`
	SrEventName       string `json:"srEventName"`
	SrEventID         string `json:"srEventId"`
}
