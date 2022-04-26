package requestdto

// GetEventsReqDto represents request body of this API
type LiveEventsReqDto struct {
	OperatorId string `json:"operatorId"`
	PartnerId  string `json:"partnerId"`
	ProviderId string `json:"providerId"`
}
