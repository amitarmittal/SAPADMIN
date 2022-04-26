package betfair

type GetMarketResultReq struct {
	EventId  string `json:"eventId"`
	MarketId string `json:"marketId"`
	SportId  string `json:"sportId"`
}

// type GetMarketResultReqs struct {
// 	GetMarketResultReqs []GetMarketResultReq `json:"getMarketResultReqs"`
// }
