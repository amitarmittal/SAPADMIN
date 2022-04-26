package dto

type UpdateOpMarketsReqDto struct {
	OperatorId string `json:"operatorId"`
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
	MarketId   string `json:"marketId"`
}
