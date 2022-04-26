package requestdto

type Market struct {
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
	MarketId   string `json:"marketId"`
}
type MarketsResultReqDto struct {
	OperatorId string   `json:"operatorId"`
	Markets    []Market `json:"markets"`
}
