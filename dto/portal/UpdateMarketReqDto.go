package dto

type UpdateMarketsReqDto struct {
	ProviderId string `json:"providerId"`
	SportId    string `json:"sportId"`
	EventId    string `json:"eventId"`
	MarketId   string `json:"marketId"`
}
