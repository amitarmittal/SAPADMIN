package commondto

type OperStatusInProviderReqDto struct {
	OperatorId string `json:"operatorId"`
	ProviderId string `json:"providerId"`
	PartnerId  string `json:"partnerId"`
	Status     string `json:"status"`
}
