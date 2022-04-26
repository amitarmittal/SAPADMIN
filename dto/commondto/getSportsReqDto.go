package commondto

//GetSportsReqDto
type GetSportsReqDto struct {
	Token      string `json:"token"`      // optional for portal users
	OperatorId string `json:"operatorId"` // optional for portal users
	PartnerId  string `json:"partnerId"`
	ProviderId string `json:"providerId"` // mandatory.
}
