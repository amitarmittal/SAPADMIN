package commondto

type ClearedOrdersRespDto struct {
	ClearedOrders []struct {
		EventTypeID         string      `json:"eventTypeId"`
		EventID             string      `json:"eventId"`
		MarketID            string      `json:"marketId"`
		SelectionID         int         `json:"selectionId"`
		Handicap            float64     `json:"handicap"`
		BetID               string      `json:"betId"`
		PlacedDate          string      `json:"placedDate"`
		PersistenceType     string      `json:"persistenceType"`
		OrderType           string      `json:"orderType"`
		Side                string      `json:"side"`
		ItemDescription     interface{} `json:"itemDescription"`
		BetOutcome          string      `json:"betOutcome"`
		PriceRequested      float64     `json:"priceRequested"`
		SettledDate         string      `json:"settledDate"`
		LastMatchedDate     string      `json:"lastMatchedDate"`
		BetCount            int         `json:"betCount"`
		Commission          float64     `json:"commission"`
		PriceMatched        float64     `json:"priceMatched"`
		PriceReduced        bool        `json:"priceReduced"`
		SizeSettled         float64     `json:"sizeSettled"`
		Profit              float64     `json:"profit"`
		SizeCancelled       float64     `json:"sizeCancelled"`
		CustomerOrderRef    string      `json:"customerOrderRef"`
		CustomerStrategyRef string      `json:"customerStrategyRef"`
	} `json:"clearedOrders"`
	MoreAvailable bool `json:"moreAvailable"`
}

type BFClearedOrdersResp struct {
	Status            string               `json:"status"`  // RS_OK or RS_ERROR
	ErrorDescription  string               `json:"message"` // Error message
	ClearedOrdersResp ClearedOrdersRespDto `json:"data"`
}
