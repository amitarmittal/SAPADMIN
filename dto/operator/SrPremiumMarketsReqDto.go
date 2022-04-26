package operatordto

type SrPremiumMarketsReqDto struct {
	OperatorId string `json:"operatorId"`
	PartnerId  string `json:"partnerId"`
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
	Token      string `json:"token"`
}
