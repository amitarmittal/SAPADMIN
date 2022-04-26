package requestdto

//GetSportsReqDto
type CompetitionsReqDto struct {
	OperatorId string `json:"operatorId"` // mandatory for feed service, optional for portal service
	PartnerId  string `json:"partnerId"`  // mandatory
	ProviderId string `json:"providerId"` // mandatory
	SportId    string `json:"sportId"`    // mandatory
}
