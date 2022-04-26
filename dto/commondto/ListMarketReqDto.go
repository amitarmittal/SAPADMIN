package commondto

type ListMarketReqDto struct {
	MarketIds       []string `json:"marketIds"`
	PriceProjection struct {
		PriceData  []interface{} `json:"priceData"`
		Virtualise bool          `json:"virtualise"`
	} `json:"priceProjection"`
}
