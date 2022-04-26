package commondto

type OperStatusInSportReqDto struct {
	OperatorId string `json:"operatorId"`
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	Status     string `json:"status"`
}
